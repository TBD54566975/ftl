package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/internal/log"
)

// NewControllerProvisioner creates a new provisioner that uses the FTL controller to provision modules
func NewControllerProvisioner(client ftlv1connect.ControllerServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[ResourceType]InMemResourceProvisionerFn{
		ResourceTypeModule: func(ctx context.Context, res *provisioner.Resource, module, _ string, step *InMemProvisioningStep) {
			defer step.Done.Store(true)

			mod, ok := res.Resource.(*provisioner.Resource_Module)
			if !ok {
				panic(fmt.Errorf("unexpected resource type: %T", res.Resource))
			}
			logger := log.FromContext(ctx)
			logger.Infof("provisioning module: %s", module)

			resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
				Schema:    mod.Module.Schema,
				Artefacts: mod.Module.Artefacts,
				Labels:    mod.Module.Labels,
			}))
			if err != nil {
				step.Err = err
			} else {
				if mod.Module.Output == nil {
					mod.Module.Output = &provisioner.ModuleResource_ModuleResourceOutput{}
				}
				mod.Module.Output.DeploymentKey = resp.Msg.DeploymentKey
			}
		},
	})
}
