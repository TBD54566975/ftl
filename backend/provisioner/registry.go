package provisioner

import (
	"context"
	"fmt"
	"strings"

	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
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
func (reg *ProvisionerRegistry) CreateDeployment(ctx context.Context, desiredModule, existingModule *schema.Module) *Deployment {
	logger := log.FromContext(ctx)
	module := desiredModule.GetName()

	deployment := &Deployment{
		Module:   desiredModule,
		Previous: existingModule,
	}

	allDesired := schema.GetProvisionedResources(desiredModule)
	allExisting := schema.GetProvisionedResources(existingModule)

	for _, binding := range reg.listBindings() {
		desired := allDesired.FilterByType(binding.Types...)
		existing := allExisting.FilterByType(binding.Types...)

		if !desired.IsEqual(existing) {
			logger.Debugf("Adding task for module %s: %s", module, binding.ID)
			deployment.Tasks = append(deployment.Tasks, &Task{
				module:     module,
				binding:    binding,
				deployment: deployment,
			})
		}
	}
	return deployment
}
