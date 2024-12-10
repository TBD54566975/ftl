package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
)

// NewControllerProvisioner creates a new provisioner that uses the FTL controller to provision modules
func NewControllerProvisioner(client ftlv1connect.ControllerServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeModule: func(ctx context.Context, moduleName string, res schema.Provisioned) (*RuntimeEvent, error) {
			logger := log.FromContext(ctx)
			logger.Debugf("Provisioning module: %s", moduleName)

			var artefacts []*ftlv1.DeploymentArtefact

			module, ok := res.(*schema.Module)
			if !ok {
				return nil, fmt.Errorf("expected module, got %T", res)
			}

			for _, artefact := range module.Metadata {
				if metadata, ok := artefact.(*schema.MetadataArtefact); ok {
					artefacts = append(artefacts, &ftlv1.DeploymentArtefact{
						Path:       metadata.Path,
						Digest:     metadata.Digest,
						Executable: metadata.Executable,
					})
				}
			}

			resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
				Schema:    module.ToProto().(*schemapb.Module),
				Artefacts: artefacts,
			}))
			if err != nil {
				return nil, fmt.Errorf("failed to create deployment: %w", err)
			}

			return &RuntimeEvent{
				Module: &schema.ModuleRuntimeDeployment{
					DeploymentKey: resp.Msg.DeploymentKey,
				},
			}, nil
		},
	})
}
