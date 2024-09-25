package deployment

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
)

// ResourceType is a type of resource used to configure provisioners
type ResourceType string

const (
	ResourceTypeUnknown  ResourceType = "unknown"
	ResourceTypePostgres ResourceType = "postgres"
	ResourceTypeMysql    ResourceType = "mysql"
)

// Provisioner is a runnable process to provision resources
type Provisioner interface {
	Provision(ctx context.Context, module string, desired []*provisioner.ResourceContext, existing []*provisioner.Resource) (string, error)
	State(ctx context.Context, token string, desired []*provisioner.Resource) (TaskState, []*provisioner.Resource, error)
}

type provisionerConfig struct {
	provisioner Provisioner
	types       []ResourceType
}

// ProvisionerRegistry contains all known resource handlers in the order they should be executed
type ProvisionerRegistry struct {
	Provisioners []*provisionerConfig
}

// Register to the registry, to be executed after all the previously added handlers
func (reg *ProvisionerRegistry) Register(handler Provisioner, types ...ResourceType) {
	reg.Provisioners = append(reg.Provisioners, &provisionerConfig{
		provisioner: handler,
		types:       types,
	})
}

// CreateDeployment to take the system to the desired state
func (reg *ProvisionerRegistry) CreateDeployment(module string, desiredResources, existingResources []*provisioner.Resource) *Deployment {
	var result []*Task

	existingByHandler := reg.groupByProvisioner(existingResources)
	desiredByHandler := reg.groupByProvisioner(desiredResources)

	for handler, desired := range desiredByHandler {
		existing := existingByHandler[handler]
		result = append(result, &Task{
			Handler:  handler,
			Desired:  desired,
			Existing: existing,
		})
	}
	return &Deployment{Tasks: result, Module: module}
}

// ExtractResources from a module schema
func ExtractResources(sch *schema.Module) ([]*provisioner.Resource, error) {
	var result []*provisioner.Resource
	for _, decl := range sch.Decls {
		if db, ok := decl.(*schema.Database); ok {
			if db.Type == "postgres" {
				result = append(result, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Postgres{},
				})
			} else if db.Type == "mysql" {
				result = append(result, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Mysql{},
				})
			} else {
				return nil, fmt.Errorf("unknown db type: %s", db.Type)
			}
		}
	}
	return result, nil
}

func (reg *ProvisionerRegistry) groupByProvisioner(resources []*provisioner.Resource) map[Provisioner][]*provisioner.Resource {
	result := map[Provisioner][]*provisioner.Resource{}
	for _, r := range resources {
		for _, cfg := range reg.Provisioners {
			for _, t := range cfg.types {
				typed := typeOf(r)
				if t == typed {
					result[cfg.provisioner] = append(result[cfg.provisioner], r)
					break
				}
			}
		}
	}
	return result
}

func typeOf(r *provisioner.Resource) ResourceType {
	if _, ok := r.Resource.(*provisioner.Resource_Mysql); ok {
		return ResourceTypeMysql
	} else if _, ok := r.Resource.(*provisioner.Resource_Postgres); ok {
		return ResourceTypePostgres
	}
	return ResourceTypeUnknown
}

// PluginProvisioner delegates provisioning to an external plugin
type PluginProvisioner struct {
	cmdCtx context.Context
	client *plugin.Plugin[provisionerconnect.ProvisionerPluginServiceClient]
}

func NewPluginProvisioner(ctx context.Context, name, dir, exe string) (*PluginProvisioner, error) {
	client, cmdCtx, err := plugin.Spawn(
		ctx,
		log.Debug,
		name,
		dir,
		exe,
		provisionerconnect.NewProvisionerPluginServiceClient,
	)
	if err != nil {
		return nil, fmt.Errorf("error spawning plugin: %w", err)
	}

	return &PluginProvisioner{
		cmdCtx: cmdCtx,
		client: client,
	}, nil
}

func (p *PluginProvisioner) Provision(ctx context.Context, module string, desired []*provisioner.ResourceContext, existing []*provisioner.Resource) (string, error) {
	resp, err := p.client.Client.Provision(ctx, connect.NewRequest(&provisioner.ProvisionRequest{
		DesiredResources:  desired,
		ExistingResources: existing,
		FtlClusterId:      "ftl",
		Module:            module,
	}))
	if err != nil {
		return "", fmt.Errorf("error calling plugin: %w", err)
	}
	if resp.Msg.Status != provisioner.ProvisionResponse_SUBMITTED {
		return resp.Msg.ProvisioningToken, nil
	}
	return "", nil
}

func (p *PluginProvisioner) State(ctx context.Context, token string, desired []*provisioner.Resource) (TaskState, []*provisioner.Resource, error) {
	resp, err := p.client.Client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
		ProvisioningToken: token,
	}))
	if err != nil {
		return "", nil, fmt.Errorf("error getting status from plugin: %w", err)
	}
	if failed, ok := resp.Msg.Status.(*provisioner.StatusResponse_Failed); ok {
		return TaskStateFailed, nil, fmt.Errorf("provisioning failed: %s", failed.Failed.ErrorMessage)
	} else if success, ok := resp.Msg.Status.(*provisioner.StatusResponse_Success); ok {
		return TaskStateDone, success.Success.UpdatedResources, nil
	}
	return TaskStateRunning, nil, nil
}

var _ Provisioner = (*PluginProvisioner)(nil)
