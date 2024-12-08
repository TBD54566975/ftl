package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
)

// NewControllerProvisioner creates a new provisioner that uses the FTL controller to provision modules
func NewControllerProvisioner(client ftlv1connect.ControllerServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeModule: func(ctx context.Context, rc *provisioner.ResourceContext, module, _ string, previous *provisioner.Resource) (*provisioner.Resource, error) {
			mod, ok := rc.Resource.Resource.(*provisioner.Resource_Module)
			if !ok {
				panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
			}
			logger := log.FromContext(ctx)
			logger.Debugf("Provisioning module: %s", module)

			for _, dep := range rc.Dependencies {
				switch r := dep.Resource.(type) {
				case *provisioner.Resource_Postgres:
					if r.Postgres == nil || r.Postgres.Output == nil {
						return nil, fmt.Errorf("postgres resource has not been provisioned")
					}

					decl, ok := findDecl(mod.Module.Schema, func(d *schemapb.Decl_Database) bool {
						return d.Database.Name == dep.ResourceId
					})
					if !ok {
						return nil, fmt.Errorf("failed to find database declaration for %s", dep.ResourceId)
					}
					decl.Database.Runtime = r.Postgres.Output
				case *provisioner.Resource_Mysql:
					if r.Mysql == nil || r.Mysql.Output == nil {
						return nil, fmt.Errorf("mysql resource has not been provisioned")
					}

					decl, ok := findDecl(mod.Module.Schema, func(d *schemapb.Decl_Database) bool {
						return d.Database.Name == dep.ResourceId
					})
					if !ok {
						return nil, fmt.Errorf("failed to find database declaration for %s", dep.ResourceId)
					}
					decl.Database.Runtime = r.Mysql.Output
				case *provisioner.Resource_Topic:
					if r.Topic == nil || r.Topic.Output == nil {
						return nil, fmt.Errorf("topic resource has not been provisioned")
					}
					decl, ok := findDecl(mod.Module.Schema, func(t *schemapb.Decl_Topic) bool {
						return t.Topic.Name == dep.ResourceId
					})
					if !ok {
						return nil, fmt.Errorf("failed to find topic declaration: %s", dep.ResourceId)
					}
					decl.Topic.Runtime = &schemapb.TopicRuntime{
						KafkaBrokers: r.Topic.Output.KafkaBrokers,
						TopicId:      r.Topic.Output.TopicId,
					}
				case *provisioner.Resource_Subscription:
					if r.Subscription == nil || r.Subscription.Output == nil {
						return nil, fmt.Errorf("subscription resource has not been provisioned")
					}
					decl, ok := findDecl(mod.Module.Schema, func(t *schemapb.Decl_Verb) bool {
						return t.Verb.Name == dep.ResourceId
					})
					if !ok {
						return nil, fmt.Errorf("failed to find subscription declaration: %s", dep.ResourceId)
					}
					decl.Verb.Runtime = &schemapb.VerbRuntime{
						KafkaBrokers: r.Subscription.Output.KafkaBrokers,
					}
				default:
				}
			}
			resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
				Schema:    mod.Module.Schema,
				Artefacts: mod.Module.Artefacts,
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

func findDecl[T any](schema *schemapb.Module, filter func(T) bool) (out T, ok bool) {
	for _, d := range schema.Decls {
		if decl, ok := d.Value.(T); ok && filter(decl) {
			return decl, true
		}
	}
	return out, false
}
