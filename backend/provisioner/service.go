package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	provproto "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/provisioner/scaling"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

// CommonProvisionerConfig is shared config between the production controller and development server.
type CommonProvisionerConfig struct {
	PluginConfigFile *os.File `name:"provisioner-plugin-config" help:"Path to the plugin configuration file." env:"FTL_PROVISIONER_PLUGIN_CONFIG_FILE"`
}

type Config struct {
	Bind               *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8893" env:"FTL_BIND"`
	ControllerEndpoint *url.URL `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	CommonProvisionerConfig
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type Service struct {
	controllerClient ftlv1connect.ControllerServiceClient
	currentResources *xsync.MapOf[string, *ResourceGraph]
	registry         *ProvisionerRegistry
}

var _ provisionerconnect.ProvisionerServiceHandler = (*Service)(nil)

func New(
	ctx context.Context,
	config Config,
	controllerClient ftlv1connect.ControllerServiceClient,
	registry *ProvisionerRegistry,
) (*Service, error) {
	resourceMap := xsync.NewMapOf[string, *ResourceGraph]()
	return &Service{
		controllerClient: controllerClient,
		currentResources: resourceMap,
		registry:         registry,
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)
	// TODO: Block deployments to make sure only one module is modified at a time
	moduleName := req.Msg.Schema.Name
	existingResources, _ := s.currentResources.Load(moduleName)
	desiredGraph, err := ExtractResources(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("error extracting resources from schema: %w", err)
	}
	if err := replaceOutputs(desiredGraph.Resources(), existingResources.Resources()); err != nil {
		return nil, err
	}

	deployment := s.registry.CreateDeployment(ctx, moduleName, desiredGraph, existingResources)
	running := true
	logger.Debugf("Running deployment for module %s", moduleName)
	for running {
		running, err = deployment.Progress(ctx)
		if err != nil {
			// TODO: Deal with failed deployments
			return nil, fmt.Errorf("error running a provisioner: %w", err)
		}
	}
	logger.Debugf("Finished deployment for module %s", moduleName)

	// update the resource state to match the resources updated in the deployment
	s.currentResources.Store(moduleName, deployment.Graph)

	deploymentKey := ""
	for _, r := range desiredGraph.Resources() {
		if mod, ok := r.Resource.(*provproto.Resource_Module); ok && mod.Module.Schema.Name == moduleName {
			deploymentKey = mod.Module.Output.DeploymentKey
			break
		}
	}

	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{
		DeploymentKey: deploymentKey,
	}), nil
}

func replaceOutputs(to []*provproto.Resource, from []*provproto.Resource) error {
	byID := map[string]*provproto.Resource{}
	for _, r := range from {
		byID[r.ResourceId] = r
	}
	for _, r := range to {
		existing := byID[r.ResourceId]
		if existing == nil {
			continue
		}
		switch r := r.Resource.(type) {
		case *provproto.Resource_Mysql:
			if mysqlFrom, ok := existing.Resource.(*provproto.Resource_Mysql); ok && mysqlFrom.Mysql != nil {
				if r.Mysql == nil {
					r.Mysql = &provproto.MysqlResource{}
				}
				r.Mysql.Output = mysqlFrom.Mysql.Output
			}
		case *provproto.Resource_Postgres:
			if postgresFrom, ok := existing.Resource.(*provproto.Resource_Postgres); ok && postgresFrom.Postgres != nil {
				if r.Postgres == nil {
					r.Postgres = &provproto.PostgresResource{}
				}
				r.Postgres.Output = postgresFrom.Postgres.Output
			}
		case *provproto.Resource_Module:
			if moduleFrom, ok := existing.Resource.(*provproto.Resource_Module); ok {
				r.Module.Output = moduleFrom.Module.Output
			}
		case *provproto.Resource_Topic:
			if topicFrom, ok := existing.Resource.(*provproto.Resource_Topic); ok && topicFrom.Topic != nil {
				if r.Topic == nil {
					r.Topic = &provproto.TopicResource{}
				}
				r.Topic.Output = topicFrom.Topic.Output
			}
		case *provproto.Resource_Subscription:
			if subscriptionFrom, ok := existing.Resource.(*provproto.Resource_Subscription); ok && subscriptionFrom.Subscription != nil {
				if r.Subscription == nil {
					r.Subscription = &provproto.SubscriptionResource{}
				}
				r.Subscription.Output = subscriptionFrom.Subscription.Output
			}
		case *provproto.Resource_Runner:
			if runnerFrom, ok := existing.Resource.(*provproto.Resource_Runner); ok && runnerFrom.Runner != nil {
				if r.Runner == nil {
					r.Runner = &provproto.RunnerResource{}
				}
				r.Runner.Output = runnerFrom.Runner.Output
			}
			// Ignore
		default:
			return fmt.Errorf("can not replace outputs for an unknown resource type %T", r)
		}
	}
	return nil
}

// Start the Provisioner. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	registry *ProvisionerRegistry,
	controllerClient ftlv1connect.ControllerServiceClient,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL provisioner")

	svc, err := New(ctx, config, controllerClient, registry)
	if err != nil {
		return err
	}
	logger.Debugf("Provisioner available at: %s", config.Bind)
	logger.Debugf("Using FTL endpoint: %s", config.ControllerEndpoint)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(provisionerconnect.NewProvisionerServiceHandler, svc),
			rpc.PProf(),
		)
	})
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error waiting for rpc.Serve: %w", err)
	}
	return nil
}

func RegistryFromConfigFile(ctx context.Context, file *os.File, controller ftlv1connect.ControllerServiceClient, scaling scaling.RunnerScaling) (*ProvisionerRegistry, error) {
	config := provisionerPluginConfig{}
	bytes, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("error reading plugin configuration from %s: %w", file.Name(), err)
	}
	if err := toml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("error parsing plugin configuration: %w", err)
	}

	registry, err := registryFromConfig(ctx, &config, controller, scaling)
	if err != nil {
		return nil, fmt.Errorf("error creating provisioner registry: %w", err)
	}

	return registry, nil
}

// Deployment client calls to ftl-controller

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	resp, err := s.controllerClient.GetArtefactDiffs(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, req *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	resp, err := s.controllerClient.ReplaceDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	resp, err := s.controllerClient.Status(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (*connect.Response[ftlv1.UpdateDeployResponse], error) {
	resp, err := s.controllerClient.UpdateDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	resp, err := s.controllerClient.UploadArtefact(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return connect.NewResponse(resp.Msg), nil
}
