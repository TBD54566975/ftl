package provisioner

import (
	"context"
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/provisioner/noop"
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

// provisionerPluginConfig is a map of provisioner name to resources it supports
type provisionerPluginConfig struct {
	// The default provisioner to use for all resources not matched here
	Default string `toml:"default"`
	Plugins []struct {
		ID        string         `toml:"id"`
		Resources []ResourceType `toml:"resources"`
	} `toml:"plugins"`
}

func (cfg *provisionerPluginConfig) Validate() error {
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

// ProvisionerBinding is a Provisioner and the types it supports
type ProvisionerBinding struct {
	Provisioner provisionerconnect.ProvisionerPluginServiceClient
	Types       []ResourceType
}

// ProvisionerRegistry contains all known resource handlers in the order they should be executed
type ProvisionerRegistry struct {
	Default      provisionerconnect.ProvisionerPluginServiceClient
	Provisioners []*ProvisionerBinding
}

func registryFromConfig(ctx context.Context, cfg *provisionerPluginConfig) (*ProvisionerRegistry, error) {
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
		return &noop.Provisioner{}, nil
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
	reg.Provisioners = append(reg.Provisioners, &ProvisionerBinding{
		Provisioner: handler,
		Types:       types,
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
func ExtractResources(msg *ftlv1.CreateDeploymentRequest) ([]*provisioner.Resource, error) {
	var result []*provisioner.Resource

	module, err := schema.ModuleFromProto(msg.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid module schema for module %s: %w", msg.Schema.Name, err)
	}

	for _, decl := range module.Decls {
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
	result = append(result, &provisioner.Resource{
		ResourceId: module.GetName(),
		Resource: &provisioner.Resource_Module{
			Module: &provisioner.ModuleResource{
				Artefacts: msg.Artefacts,
				Schema:    msg.Schema,
			},
		},
	})

	return result, nil
}

func (reg *ProvisionerRegistry) groupByProvisioner(resources []*provisioner.Resource) map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource {
	result := map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource{}
	for _, r := range resources {
		found := false
		for _, cfg := range reg.Provisioners {
			for _, t := range cfg.Types {
				typed := typeOf(r)
				if t == typed {
					result[cfg.Provisioner] = append(result[cfg.Provisioner], r)
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
