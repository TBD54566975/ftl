package provisioner

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/backend/provisioner/deployment"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type Config struct {
	Bind               *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8894" env:"FTL_PROVISIONER_BIND"`
	IngressBind        *url.URL `help:"Socket to bind to for ingress." default:"http://127.0.0.1:8893" env:"FTL_PROVISIONER_INGRESS_BIND"`
	Advertise          *url.URL `help:"Endpoint the Provisioner should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_PROVISIONER_ADVERTISE"`
	ControllerEndpoint *url.URL `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
	if c.Advertise == nil {
		c.Advertise = c.Bind
	}
}

type Service struct {
	controllerClient ftlv1connect.ControllerServiceClient
	// TODO: Store in a resource graph
	currentResources map[string][]*provisioner.Resource
	registry         deployment.ProvisionerRegistry
}

var _ provisionerconnect.ProvisionerServiceHandler = (*Service)(nil)

func New(ctx context.Context, config Config, controllerClient ftlv1connect.ControllerServiceClient, devel bool) (*Service, error) {
	return &Service{
		controllerClient: controllerClient,
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	// TODO: Block deployments to make sure only one module is modified at a time
	moduleName := req.Msg.Schema.Name
	module, err := schema.ModuleFromProto(req.Msg.Schema)
	if err != nil {
		return nil, fmt.Errorf("invalid module schema for module %s: %w", moduleName, err)
	}

	existingResources := s.currentResources[moduleName]
	desiredResources, err := deployment.ExtractResources(module)
	if err != nil {
		return nil, fmt.Errorf("error extracting resources from schema: %w", err)
	}
	if err := replaceOutputs(desiredResources, existingResources); err != nil {
		return nil, err
	}

	deployment := s.registry.CreateDeployment(moduleName, desiredResources, existingResources)
	running := true
	for running {
		running, err = deployment.Progress(ctx)
		if err != nil {
			// TODO: Deal with failed deployments
			return nil, fmt.Errorf("error running a provisioner: %w", err)
		}
	}

	// TODO: manage multiple deployments properly. Extract as a provisioner plugin
	response, err := s.controllerClient.CreateDeployment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}

	s.currentResources[moduleName] = desiredResources

	return response, nil
}

func replaceOutputs(to []*provisioner.Resource, from []*provisioner.Resource) error {
	byID := map[string]*provisioner.Resource{}
	for _, r := range from {
		byID[r.ResourceId] = r
	}
	for _, r := range to {
		existing := byID[r.ResourceId]
		if existing == nil {
			continue
		}
		if mysqlTo, ok := r.Resource.(*provisioner.Resource_Mysql); ok {
			if myslqFrom, ok := existing.Resource.(*provisioner.Resource_Mysql); ok {
				mysqlTo.Mysql.Output = myslqFrom.Mysql.Output
			}
		} else if postgresTo, ok := r.Resource.(*provisioner.Resource_Postgres); ok {
			if postgresFrom, ok := existing.Resource.(*provisioner.Resource_Postgres); ok {
				postgresTo.Postgres.Output = postgresFrom.Postgres.Output
			}
		} else {
			return fmt.Errorf("can not replace outputs for an unknown resource type %T", r)
		}
	}
	return nil
}

// Start the Provisioner. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config, devel bool) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL provisioner")

	controllerClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, config.ControllerEndpoint.String(), log.Error)

	svc, err := New(ctx, config, controllerClient, devel)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)
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

// Deployment client calls to ftl-controller

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	resp, err := s.controllerClient.GetArtefactDiffs(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, req *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	resp, err := s.controllerClient.ReplaceDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	resp, err := s.controllerClient.Status(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (*connect.Response[ftlv1.UpdateDeployResponse], error) {
	resp, err := s.controllerClient.UpdateDeploy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	resp, err := s.controllerClient.UploadArtefact(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) GetSchema(ctx context.Context, req *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	resp, err := s.controllerClient.GetSchema(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return resp, nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], to *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	stream, err := s.controllerClient.PullSchema(ctx, req)
	if err != nil {
		return fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	defer stream.Close()
	for stream.Receive() {
		if err := stream.Err(); err != nil {
			return fmt.Errorf("call to ftl-controller failed: %w", err)
		}
		if err := to.Send(stream.Msg()); err != nil {
			return fmt.Errorf("call to ftl-controller failed: %w", err)
		}
	}
	return nil
}
