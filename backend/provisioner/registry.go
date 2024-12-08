package provisioner

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/provisioner/scaling"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

// provisionerPluginConfig is a map of provisioner name to resources it supports
type provisionerPluginConfig struct {
	// The default provisioner to use for all resources not matched here
	Default string `toml:"default"`
	Plugins []struct {
		ID        string                `toml:"id"`
		Resources []schema.ResourceType `toml:"resources"`
	} `toml:"plugins"`
}

func (cfg *provisionerPluginConfig) Validate() error {
	registeredResources := map[schema.ResourceType]bool{}
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
	Types       []schema.ResourceType
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
	Bindings []*ProvisionerBinding
}

// listBindings in the order they should be executed
func (reg *ProvisionerRegistry) listBindings() []*ProvisionerBinding {
	result := []*ProvisionerBinding{}
	result = append(result, reg.Bindings...)
	return result
}

func registryFromConfig(ctx context.Context, cfg *provisionerPluginConfig, controller ftlv1connect.ControllerServiceClient, runnerScaling scaling.RunnerScaling) (*ProvisionerRegistry, error) {
	logger := log.FromContext(ctx)
	result := &ProvisionerRegistry{}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating provisioner config: %w", err)
	}
	for _, plugin := range cfg.Plugins {
		provisioner, err := provisionerIDToProvisioner(ctx, plugin.ID, controller, runnerScaling)
		if err != nil {
			return nil, err
		}
		binding := result.Register(plugin.ID, provisioner, plugin.Resources...)
		logger.Debugf("Registered provisioner %s", binding)
	}
	return result, nil
}

func provisionerIDToProvisioner(ctx context.Context, id string, controller ftlv1connect.ControllerServiceClient, scaling scaling.RunnerScaling) (provisionerconnect.ProvisionerPluginServiceClient, error) {
	switch id {
	case "controller":
		return NewControllerProvisioner(controller), nil
	case "kubernetes":
		// TODO: move this into a plugin
		return NewRunnerScalingProvisioner(scaling), nil
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
func (reg *ProvisionerRegistry) Register(id string, handler provisionerconnect.ProvisionerPluginServiceClient, types ...schema.ResourceType) *ProvisionerBinding {
	binding := &ProvisionerBinding{
		Provisioner: handler,
		Types:       types,
		ID:          id,
	}
	reg.Bindings = append(reg.Bindings, binding)
	return binding
}

// CreateDeployment to take the system to the desired state
func (reg *ProvisionerRegistry) CreateDeployment(ctx context.Context, module string, desiredResources, existingResources *ResourceGraph) *Deployment {
	logger := log.FromContext(ctx)

	deployment := &Deployment{
		Module: module,
		Graph:  desiredResources,
	}

	for _, binding := range reg.listBindings() {
		desired := getTypes(desiredResources.Resources(), binding.Types)
		existing := getTypes(existingResources.Resources(), binding.Types)

		if !resourcesEqual(desired, existing) {
			logger.Debugf("Adding task for module %s: %s", module, binding)
			deployment.Tasks = append(deployment.Tasks, &Task{
				module:     module,
				binding:    binding,
				deployment: deployment,
				desired:    desiredResources.WithDirectDependencies(desired),
			})
		} else {
			logger.Debugf("Skipping task for module %s with provisioner %s", module, binding.ID)
		}
	}
	return deployment
}

func getTypes(resources []*provisioner.Resource, types []schema.ResourceType) []*provisioner.Resource {
	result := []*provisioner.Resource{}
	for _, r := range resources {
		for _, t := range types {
			if TypeOf(r) == t {
				result = append(result, r)
			}
		}
	}
	return result
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
			&schemapb.DatabaseRuntime{},
			&provisioner.ModuleResource_ModuleResourceOutput{},
			&provisioner.SqlMigrationResource_SqlMigrationResourceOutput{},
			&provisioner.TopicResource_TopicResourceOutput{},
			&provisioner.SubscriptionResource_SubscriptionResourceOutput{},
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
	edges := make([]*ResourceEdge, 0)
	for _, decl := range module.Decls {
		switch decl := decl.(type) {
		case *schema.Database:
			switch decl.Type {
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
				return nil, fmt.Errorf("unknown db type: %s", decl.Type)
			}

			if migration, ok := slices.FindVariant[*schema.MetadataSQLMigration](decl.Metadata); ok {
				id := decl.GetName() + "-migration-" + migration.Digest
				deps = append(deps, &provisioner.Resource{
					ResourceId: id,
					Resource:   &provisioner.Resource_SqlMigration{SqlMigration: &provisioner.SqlMigrationResource{Digest: migration.Digest}},
				})
				edges = append(edges, &ResourceEdge{
					from: id,
					to:   decl.GetName(),
				})
			}
		case *schema.Topic:
			deps = append(deps, &provisioner.Resource{
				ResourceId: decl.GetName(),
				Resource:   &provisioner.Resource_Topic{},
			})
		case *schema.Verb:
			subscriber, ok := slices.FindVariant[*schema.MetadataSubscriber](decl.Metadata)
			if !ok {
				continue
			}
			deps = append(deps, &provisioner.Resource{
				ResourceId: decl.GetName(),
				Resource: &provisioner.Resource_Subscription{
					Subscription: &provisioner.SubscriptionResource{
						Topic: &schemapb.Ref{
							Module: subscriber.Topic.Module,
							Name:   subscriber.Topic.Name,
						},
					},
				},
			})
		case *schema.Config, *schema.Data, *schema.Enum, *schema.Secret, *schema.TypeAlias:
		}
	}

	root := &provisioner.Resource{
		ResourceId: module.GetName(),
		Resource: &provisioner.Resource_Module{
			Module: &provisioner.ModuleResource{
				Schema:    msg.Schema,
				Artefacts: msg.Artefacts,
			},
		},
	}
	for _, dep := range deps {
		edges = append(edges, &ResourceEdge{
			from: root.ResourceId,
			to:   dep.ResourceId,
		})
	}
	digests := ""
	orderedDigests := []string{}
	for _, d := range msg.Artefacts {
		orderedDigests = append(orderedDigests, d.Digest)
	}
	slices.Sort(orderedDigests)
	for _, d := range orderedDigests {
		digests += d
	}

	// Hack, we just use the artifact digests to create a unique runner resource
	hash := sha256.Sum([]byte(digests))

	runnerResource := &provisioner.Resource{
		ResourceId: module.GetName() + "-" + hash.String() + "-runner",
		Resource: &provisioner.Resource_Runner{
			Runner: &provisioner.RunnerResource{},
		},
	}
	edges = append(edges, &ResourceEdge{
		from: runnerResource.ResourceId,
		to:   root.ResourceId,
	})
	result := &ResourceGraph{
		nodes: append(deps, root, runnerResource),
		edges: edges,
	}

	return result, nil
}
