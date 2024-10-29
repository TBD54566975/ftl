package provisioner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
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
	ID          string
	Types       []ResourceType
}

func (p ProvisionerBinding) String() string {
	types := []string{}
	for _, t := range p.Types {
		types = append(types, string(t))
	}
	return fmt.Sprintf("%s (%s)", p.ID, strings.Join(types, ","))
}

// ProvisionerRegistry contains all known resource handlers in the order they should be executed
type ProvisionerRegistry struct {
	Default      *ProvisionerBinding
	Provisioners []*ProvisionerBinding
}

// listProvisioners in the order they should be executed
func (reg *ProvisionerRegistry) listProvisioners() []*ProvisionerBinding {
	result := []*ProvisionerBinding{}
	if reg.Default != nil {
		result = append(result, reg.Default)
	}
	result = append(result, reg.Provisioners...)
	return result
}

func registryFromConfig(ctx context.Context, cfg *provisionerPluginConfig, controller ftlv1connect.ControllerServiceClient) (*ProvisionerRegistry, error) {
	logger := log.FromContext(ctx)
	var def provisionerconnect.ProvisionerPluginServiceClient
	if cfg.Default != "" {
		d, err := provisionerIDToProvisioner(ctx, cfg.Default, controller)
		if err != nil {
			return nil, err
		}
		def = d
	}
	result := &ProvisionerRegistry{Default: &ProvisionerBinding{Provisioner: def, ID: cfg.Default}}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating provisioner config: %w", err)
	}
	for _, plugin := range cfg.Plugins {
		provisioner, err := provisionerIDToProvisioner(ctx, plugin.ID, controller)
		if err != nil {
			return nil, err
		}
		binding := result.Register(plugin.ID, provisioner, plugin.Resources...)
		logger.Debugf("Registered provisioner %s", binding)
	}
	return result, nil
}

func provisionerIDToProvisioner(ctx context.Context, id string, controller ftlv1connect.ControllerServiceClient) (provisionerconnect.ProvisionerPluginServiceClient, error) {
	switch id {
	case "controller":
		return NewControllerProvisioner(controller), nil
	case "noop":
		return &NoopProvisioner{}, nil
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
func (reg *ProvisionerRegistry) Register(id string, handler provisionerconnect.ProvisionerPluginServiceClient, types ...ResourceType) *ProvisionerBinding {
	binding := &ProvisionerBinding{
		Provisioner: handler,
		Types:       types,
		ID:          id,
	}
	reg.Provisioners = append(reg.Provisioners, binding)
	return binding
}

// CreateDeployment to take the system to the desired state
func (reg *ProvisionerRegistry) CreateDeployment(ctx context.Context, module string, desiredResources, existingResources *ResourceGraph) *Deployment {
	logger := log.FromContext(ctx)

	existingByHandler := reg.groupByProvisioner(existingResources.Resources())
	desiredByHandler := reg.groupByProvisioner(desiredResources.Resources())

	deployment := &Deployment{
		Module: module,
		Graph:  desiredResources,
	}

	for _, binding := range reg.listProvisioners() {
		desired := desiredByHandler[binding.Provisioner]
		existing := existingByHandler[binding.Provisioner]

		if !resourcesEqual(desired, existing) {
			logger.Debugf("Adding task for module %s: %s", module, binding)
			deployment.Tasks = append(deployment.Tasks, &Task{
				module:     module,
				handler:    binding.Provisioner,
				deployment: deployment,
				desired:    desiredResources.WithDirectDependencies(desired),
			})
		} else {
			logger.Debugf("Skipping task for module %s with provisioner %s", module, binding.ID)
		}
	}
	return deployment
}

func resourcesEqual(desired, existing []*provisioner.Resource) bool {
	if len(desired) != len(existing) {
		return false
	}
	// sort by resource id
	sort.Slice(desired, func(i, j int) bool {
		return desired[i].ResourceId < desired[j].ResourceId
	})
	sort.Slice(existing, func(i, j int) bool {
		return existing[i].ResourceId < existing[j].ResourceId
	})
	// check each resource
	for i := range desired {
		if !resourceEqual(desired[i], existing[i]) {
			return false
		}
	}
	return true
}

func resourceEqual(desired, existing *provisioner.Resource) bool {
	return cmp.Equal(desired, existing,
		protocmp.Transform(),
		protocmp.IgnoreMessages(
			&provisioner.MysqlResource_MysqlResourceOutput{},
			&provisioner.PostgresResource_PostgresResourceOutput{},
			&provisioner.ModuleResource_ModuleResourceOutput{},
		),
	)
}

// ExtractResources from a module schema
func ExtractResources(msg *ftlv1.CreateDeploymentRequest) (*ResourceGraph, error) {
	var deps []*provisioner.Resource

	module, err := schema.ModuleFromProto(msg.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid module schema for module %s: %w", msg.Schema.Name, err)
	}

	for _, decl := range module.Decls {
		if db, ok := decl.(*schema.Database); ok {
			switch db.Type {
			case "postgres":
				deps = append(deps, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Postgres{},
				})
			case "mysql":
				deps = append(deps, &provisioner.Resource{
					ResourceId: decl.GetName(),
					Resource:   &provisioner.Resource_Mysql{},
				})
			default:
				return nil, fmt.Errorf("unknown db type: %s", db.Type)
			}
		}
	}

	root := &provisioner.Resource{
		ResourceId: module.GetName(),
		Resource: &provisioner.Resource_Module{
			Module: &provisioner.ModuleResource{
				Schema:    msg.Schema,
				Artefacts: msg.Artefacts,
				Labels:    msg.Labels,
			},
		},
	}
	edges := make([]*ResourceEdge, len(deps))
	for i, dep := range deps {
		edges[i] = &ResourceEdge{
			from: root.ResourceId,
			to:   dep.ResourceId,
		}
	}

	result := &ResourceGraph{
		nodes: append(deps, root),
		edges: edges,
	}

	return result, nil
}

func (reg *ProvisionerRegistry) groupByProvisioner(resources []*provisioner.Resource) map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource {
	result := map[provisionerconnect.ProvisionerPluginServiceClient][]*provisioner.Resource{}
	for _, r := range resources {
		found := false
		for _, cfg := range reg.Provisioners {
			for _, t := range cfg.Types {
				typed := TypeOf(r)
				if t == typed {
					result[cfg.Provisioner] = append(result[cfg.Provisioner], r)
					found = true
					break
				}
			}
		}
		if !found && reg.Default != nil {
			result[reg.Default.Provisioner] = append(result[reg.Default.Provisioner], r)
		}
	}
	return result
}
