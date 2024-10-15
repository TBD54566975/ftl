package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/internal/log"
)

// NewControllerProvisioner creates a new provisioner that uses the FTL controller to provision modules
func NewControllerProvisioner(client ftlv1connect.ControllerServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[ResourceType]InMemResourceProvisionerFn{
		ResourceTypeModule: func(ctx context.Context, rc *provisioner.ResourceContext, module, _ string) (*provisioner.Resource, error) {
			mod, ok := rc.Resource.Resource.(*provisioner.Resource_Module)
			if !ok {
				panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
			}
			logger := log.FromContext(ctx)
			logger.Infof("provisioning module: %s", module)

			for _, dep := range rc.Dependencies {
				if psql, ok := dep.Resource.(*provisioner.Resource_Postgres); ok {
					if psql.Postgres == nil || psql.Postgres.Output == nil {
						return nil, fmt.Errorf("postgres resource has not been provisioned")
					}

					decl, err := findDBDecl(dep.ResourceId, mod.Module.Schema)
					if err != nil {
						return nil, fmt.Errorf("failed to find database declaration: %w", err)
					}
					decl.Runtime = &schemapb.DatabaseRuntime{
						Dsn: psql.Postgres.Output.WriteDsn,
					}
				}
			}

			resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
				Schema:    mod.Module.Schema,
				Artefacts: mod.Module.Artefacts,
				Labels:    mod.Module.Labels,
			}))
			if err != nil {
				return nil, fmt.Errorf("failed to create deployment: %w", err)
			}
			if mod.Module.Output == nil {
				mod.Module.Output = &provisioner.ModuleResource_ModuleResourceOutput{}
			}
			mod.Module.Output.DeploymentKey = resp.Msg.DeploymentKey
			return rc.Resource, nil
		},
	})
}

func findDBDecl(id string, schema *schemapb.Module) (*schemapb.Database, error) {
	for _, d := range schema.Decls {
		if db, ok := d.Value.(*schemapb.Decl_Database); ok {
			if db.Database.Name == id {
				return db.Database, nil
			}
		}
	}
	return nil, fmt.Errorf("database %s not found", id)
}
