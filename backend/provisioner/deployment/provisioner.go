package deployment

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
)

// ResourceType is a type of resource used to configure provisioners
type ResourceType string

const (
	ResourceTypeUnknown  ResourceType = "unknown"
	ResourceTypePostgres ResourceType = "postgres"
	ResourceTypeMysql    ResourceType = "mysql"
)

// ProvisionerPluginConfig is a map of provisioner name to resources it supports
type ProvisionerPluginConfig struct {
	// The default provisioner to use for all resources not matched here
	Default string `toml:"default"`
	Plugins []struct {
		ID        string         `toml:"id"`
		Resources []ResourceType `toml:"resources"`
	} `toml:"plugins"`
}

func (cfg *ProvisionerPluginConfig) Validate() error {
	registeredResources := map[ResourceType]bool{}
	for _, plugin := range cfg.Plugins {
		for _, r := range plugin.Resources {
			if registeredResources[r] {
				return fmt.Errorf("resource type %s is already registered. Trying to re-register for %s", r, plugin.ID)
			}
			registeredResources[r] = true
		}
	}
	return nil
}

type provisionerConfig struct {
	provisioner provisionerconnect.ProvisionerPluginServiceClient
	types       []ResourceType
}

// ProvisionerRegistry contains all known resource handlers in the order they should be executed
type ProvisionerRegistry struct {
	Default      provisionerconnect.ProvisionerPluginServiceClient
	Provisioners []*provisionerConfig
}

func NewProvisionerRegistry(ctx context.Context, cfg *ProvisionerPluginConfig) (*ProvisionerRegistry, error) {
	def, err := provisionerIDToProvisioner(ctx, cfg.Default)
	if err != nil {
		return nil, err
	}
	result := &ProvisionerRegistry{Default: def}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating provisioner config: %w", err)
	}
	for _, plugin := range cfg.Plugins {
		provisioner, err := provisionerIDToProvisioner(ctx, plugin.ID)
		if err != nil {
			return nil, err
		}
		result.Register(provisioner, plugin.Resources...)
	}
	return result, nil
}

func provisionerIDToProvisioner(ctx context.Context, id string) (provisionerconnect.ProvisionerPluginServiceClient, error) {
	switch id {
	case "noop":
		return &NoopProvisioner{}, nil
	case "dev":
		// TODO: Wire in settings from ftl serve
		return NewDevProvisioner("postgres:15.8", 15432), nil
	default:
		plugin, _, err := plugin.Spawn(
			ctx,
			log.Debug,
			"ftl-provisioner-"+id,
			".",
			"ftl-provisioner-"+id,
			provisionerconnect.NewProvisionerPluginServiceClient,
		)
		if err != nil {
			return nil, fmt.Errorf("error spawning plugin: %w", err)
		}

		return plugin.Client, nil
	}
}

// Register to the registry, to be executed after all the previously added handlers
func (reg *ProvisionerRegistry) Register(handler provisionerconnect.ProvisionerPluginServiceClient, types ...ResourceType) {
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
			module:   module,
			handler:  handler,
			desired:  desired,
			existing: existing,
		})
	}
	return &Deployment{Tasks: result, Module: module}
}

// ExtractResources from a module schema
func ExtractResources(sch *schema.Module) ([]*provisioner.Resource, error) {
	var result []*provisioner.Resource
	for _, decl := range sch.Decls {
		if db, ok := decl.(*schema.Database); ok {
			switch db.Type {
			case "postgres":
				result = append(result, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Postgres{},
				})
			case "mysql":
				result = append(result, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Mysql{},
				})
			default:
				return nil, fmt.Errorf("unknown db type: %s", db.Type)
			}
		}
	}
	return result, nil
}

func (reg *ProvisionerRegistry) groupByProvisioner(resources []*provisioner.Resource) map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource {
	result := map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource{}
	for _, r := range resources {
		found := false
		for _, cfg := range reg.Provisioners {
			for _, t := range cfg.types {
				typed := typeOf(r)
				if t == typed {
					result[cfg.provisioner] = append(result[cfg.provisioner], r)
					found = true
					break
				}
			}
		}
		if !found {
			result[reg.Default] = append(result[reg.Default], r)
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
