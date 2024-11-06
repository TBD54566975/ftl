package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	cf "github.com/awslabs/goformation/v7/cloudformation/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
)

const (
	PropertyDBReadEndpoint  = "db:read_endpoint"
	PropertyDBWriteEndpoint = "db:write_endpoint"
	PropertyMasterUserARN   = "db:maser_user_secret_arn"
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

	return ctx, &CloudformationProvisioner{client: client, secrets: secrets, confg: &config}, nil
}

func (c *CloudformationProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (c *CloudformationProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	res, updated, err := c.createChangeSet(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	if !updated {
		return connect.NewResponse(&provisioner.ProvisionResponse{
			// even if there are no changes, return the stack id so that any resource outputs can be populated
			Status:            provisioner.ProvisionResponse_SUBMITTED,
			ProvisioningToken: *res.StackId,
		}), nil
	}
	_, err = c.client.ExecuteChangeSet(ctx, &cloudformation.ExecuteChangeSetInput{
		ChangeSetName: res.Id,
		StackName:     res.StackId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute change-set: %w", err)
	}

	return connect.NewResponse(&provisioner.ProvisionResponse{
		Status:            provisioner.ProvisionResponse_SUBMITTED,
		ProvisioningToken: *res.StackId,
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
		if err := c.resourceToCF(req.FtlClusterId, req.Module, template, resourceCtx.Resource); err != nil {
			return "", err
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

func (c *CloudformationProvisioner) resourceToCF(cluster, module string, template *goformation.Template, resource *provisioner.Resource) error {
	if _, ok := resource.Resource.(*provisioner.Resource_Postgres); ok {
		clusterID := cloudformationResourceID(resource.ResourceId, "cluster")
		instanceID := cloudformationResourceID(resource.ResourceId, "instance")
		template.Resources[clusterID] = &rds.DBCluster{
			Engine:                   ptr("aurora-postgresql"),
			MasterUsername:           ptr("root"),
			ManageMasterUserPassword: ptr(true),
			DBSubnetGroupName:        ptr(c.confg.DatabaseSubnetGroupARN),
			VpcSecurityGroupIds:      []string{c.confg.DatabaseSecurityGroup},
			EngineMode:               ptr("provisioned"),
			Port:                     ptr(5432),
			ServerlessV2ScalingConfiguration: &rds.DBCluster_ServerlessV2ScalingConfiguration{
				MinCapacity: ptr(0.5),
				MaxCapacity: ptr(10.0),
			},
			Tags: ftlTags(cluster, module),
		}
		template.Resources[instanceID] = &rds.DBInstance{
			Engine:              ptr("aurora-postgresql"),
			DBInstanceClass:     ptr("db.serverless"),
			DBClusterIdentifier: ptr(goformation.Ref(clusterID)),
			Tags:                ftlTags(cluster, module),
		}
		addOutput(template.Outputs, goformation.GetAtt(clusterID, "Endpoint.Address"), &CloudformationOutputKey{
			ResourceID:   resource.ResourceId,
			PropertyName: PropertyDBWriteEndpoint,
		})
		addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
			ResourceID:   resource.ResourceId,
			PropertyName: PropertyDBReadEndpoint,
		})
		addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
			ResourceID:   resource.ResourceId,
			PropertyName: PropertyMasterUserARN,
		})
		return nil
	}
	return errors.New("unsupported resource type")
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
