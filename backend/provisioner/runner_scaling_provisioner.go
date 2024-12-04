package provisioner

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	_ "github.com/go-sql-driver/mysql"

	provisioner "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/provisioner/scaling"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

// NewRunnerScalingProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewRunnerScalingProvisioner(runners scaling.RunnerScaling) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeRunner: provisionRunner(runners),
	})
}

func provisionRunner(scaling scaling.RunnerScaling) InMemResourceProvisionerFn {
	return func(ctx context.Context, rc *provisioner.ResourceContext, module, id string, previous *provisioner.Resource) (*provisioner.Resource, error) {
		logger := log.FromContext(ctx)
		runner, ok := rc.Resource.Resource.(*provisioner.Resource_Runner)
		if !ok {
			panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
		}
		deployment := ""
		var sch *schemapb.Module
		for _, dep := range rc.Dependencies {
			switch mod := dep.Resource.(type) {
			case *provisioner.Resource_Module:
				deployment = mod.Module.Output.DeploymentKey
				sch = mod.Module.Schema
			default:
			}
		}
		if deployment == "" {
			return rc.Resource, fmt.Errorf("failed to find deployment for runner")
		}
		schema, err := schema.ModuleFromProto(sch)
		if err != nil {
			return nil, fmt.Errorf("failed to parse schema: %w", err)
		}
		logger.Debugf("provisioning runner: %s.%s for deployment %s", module, id, deployment)
		err = scaling.StartDeployment(ctx, module, deployment, schema)
		if err != nil {
			logger.Infof("failed to start deployment: %v", err)
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}
		endpoint, err := scaling.GetEndpointForDeployment(ctx, module, deployment)
		if err != nil || !endpoint.Ok() {
			return nil, fmt.Errorf("failed to get endpoint for deployment: %w", err)
		}
		ep := endpoint.MustGet()
		endpointURI := ep.String()

		runnerClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpointURI, log.Error)
		// TODO: a proper timeout
		timeout := time.After(1 * time.Minute)
		for {
			_, err := runnerClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
			if err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled %w", ctx.Err())
			case <-timeout:
				return nil, fmt.Errorf("timed out waiting for runner to be ready")
			case <-time.After(time.Millisecond * 100):
			}
		}

		runner.Runner.Output = &provisioner.RunnerResource_RunnerResourceOutput{
			RunnerUri:     endpointURI,
			DeploymentKey: deployment,
		}
		if previous != nil && previous.GetRunner().GetOutput().GetDeploymentKey() != deployment {
			logger.Debugf("terminating previous deployment: %s", previous.GetRunner().GetOutput().GetDeploymentKey())
			err := scaling.TerminateDeployment(ctx, module, previous.GetRunner().GetOutput().GetDeploymentKey())
			if err != nil {
				logger.Errorf(err, "failed to terminate previous deployment")
			}
		}
		schemaClient := rpc.ClientFromContext[ftlv1connect.SchemaServiceClient](ctx)

		logger.Infof("updating module runtime for %s with endpoint %s", module, endpointURI)
		_, err = schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Deployment: deployment, Event: &schemapb.ModuleRuntimeEvent{Value: &schemapb.ModuleRuntimeEvent_ModuleRuntimeDeployment{ModuleRuntimeDeployment: &schemapb.ModuleRuntimeDeployment{DeploymentKey: deployment, Endpoint: endpointURI}}}}))
		if err != nil {
			return nil, fmt.Errorf("failed to update module runtime: %w", err)
		}
		return rc.Resource, nil
	}
}
