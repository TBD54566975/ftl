package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	cf "github.com/awslabs/goformation/v7/cloudformation/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
)

const (
	PropertyPsqlReadEndpoint   = "psql:read_endpoint"
	PropertyPsqlWriteEndpoint  = "psql:write_endpoint"
	PropertyPsqlMasterUserARN  = "psql:master_user_secret_arn"
	PropertyMySQLReadEndpoint  = "mysql:read_endpoint"
	PropertyMySQLWriteEndpoint = "mysql:write_endpoint"
	PropertyMySQLMasterUserARN = "mysql:master_user_secret_arn"
)

type Config struct {
	DatabaseSubnetGroupARN string `help:"ARN for the subnet group to be used to create Databases in" env:"FTL_PROVISIONER_CF_DB_SUBNET_GROUP"`
	// TODO: remove this once we have module specific security groups
	DatabaseSecurityGroup string `help:"SG for databases" env:"FTL_PROVISIONER_CF_DB_SECURITY_GROUP"`
}

type CloudformationProvisioner struct {
	client  *cloudformation.Client
	secrets *secretsmanager.Client
	confg   *Config

	running *xsync.MapOf[string, *task]
}

var _ provisionerconnect.ProvisionerPluginServiceHandler = (*CloudformationProvisioner)(nil)

func NewCloudformationProvisioner(ctx context.Context, config Config) (context.Context, *CloudformationProvisioner, error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cloudformation client: %w", err)
	}
	secrets, err := createSecretsClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create secretsmanager client: %w", err)
	}

	return ctx, &CloudformationProvisioner{
		client:  client,
		secrets: secrets,
		confg:   &config,
		running: xsync.NewMapOf[string, *task](),
	}, nil
}

func (c *CloudformationProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (c *CloudformationProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	logger := log.FromContext(ctx)

	res, updated, err := c.createChangeSet(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	token := *res.StackId
	changeSetID := *res.Id

	if !updated {
		return connect.NewResponse(&provisioner.ProvisionResponse{
			// even if there are no changes, return the stack id so that any resource outputs can be populated
			Status:            provisioner.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
			ProvisioningToken: token,
		}), nil
	}

	task := &task{stackID: token}
	if _, ok := c.running.LoadOrStore(token, task); ok {
		return nil, fmt.Errorf("provisioner already running: %s", token)
	}
	logger.Debugf("Starting task for module %s: %s (%s)", req.Msg.Module, token, changeSetID)
	task.Start(ctx, c.client, c.secrets, changeSetID)
	return connect.NewResponse(&provisioner.ProvisionResponse{
		Status:            provisioner.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
		ProvisioningToken: token,
	}), nil
}

func (c *CloudformationProvisioner) createChangeSet(ctx context.Context, req *provisioner.ProvisionRequest) (*cloudformation.CreateChangeSetOutput, bool, error) {
	stack := stackName(req)
	changeSet := generateChangeSetName(stack)
	templateStr, err := c.createTemplate(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create cloudformation template: %w", err)
	}
	if err := ensureStackExists(ctx, c.client, stack); err != nil {
		return nil, false, fmt.Errorf("failed to verify the stack exists: %w", err)
	}

	res, err := c.client.CreateChangeSet(ctx, &cloudformation.CreateChangeSetInput{
		StackName:     &stack,
		ChangeSetName: &changeSet,
		TemplateBody:  &templateStr,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to create change-set: %w", err)
	}
	updated, err := waitChangeSetReady(ctx, c.client, changeSet, stack)
	if err != nil {
		return nil, false, fmt.Errorf("failed to wait for change-set to become ready: %w", err)
	}
	return res, updated, nil
}

func stackName(req *provisioner.ProvisionRequest) string {
	return sanitize(req.FtlClusterId) + "-" + sanitize(req.Module)
}

func generateChangeSetName(stack string) string {
	return sanitize(stack) + strconv.FormatInt(time.Now().Unix(), 10)
}

func (c *CloudformationProvisioner) createTemplate(req *provisioner.ProvisionRequest) (string, error) {
	template := goformation.NewTemplate()
	for _, resourceCtx := range req.DesiredResources {
		var templater ResourceTemplater
		switch resourceCtx.Resource.Resource.(type) {
		case *provisioner.Resource_Postgres:
			templater = &PostgresTemplater{
				resourceID: resourceCtx.Resource.ResourceId,
				cluster:    req.FtlClusterId,
				module:     req.Module,
				config:     c.confg,
			}
		case *provisioner.Resource_Mysql:
			templater = &MySQLTemplater{
				resourceID: resourceCtx.Resource.ResourceId,
				cluster:    req.FtlClusterId,
				module:     req.Module,
				config:     c.confg,
			}
		default:
			continue
		}

		if err := templater.AddToTemplate(template); err != nil {
			return "", fmt.Errorf("failed to add resource to template: %w", err)
		}
	}
	// Stack can not be empty, insert a null resource to keep the stack around
	if len(req.DesiredResources) == 0 {
		template.Resources["NullResource"] = &cf.WaitConditionHandle{}
	}

	bytes, err := template.JSON()
	if err != nil {
		return "", fmt.Errorf("failed to create cloudformation template: %w", err)
	}
	return string(bytes), nil
}

// ResourceTemplater interface for different resource types
type ResourceTemplater interface {
	AddToTemplate(tmpl *goformation.Template) error
}

func ftlTags(cluster, module string) []tags.Tag {
	return []tags.Tag{{
		Key:   "ftl:module",
		Value: module,
	}, {
		Key:   "ftl:cluster",
		Value: cluster,
	}}
}

func cloudformationResourceID(strs ...string) string {
	caser := cases.Title(language.English)
	var buffer bytes.Buffer

	for _, s := range strs {
		buffer.WriteString(caser.String(s))
	}
	return buffer.String()
}

func sanitize(name string) string {
	// just keep alpha numeric chars
	s := []byte(name)
	j := 0
	for _, b := range s {
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			s[j] = b
			j++
		}
	}
	return string(s[:j])
}

func main() {
	plugin.Start(
		context.Background(),
		"ftl-provisioner-cloudformation",
		NewCloudformationProvisioner,
		"",
		provisionerconnect.NewProvisionerPluginServiceHandler,
	)
}

func ptr[T any](s T) *T { return &s }
