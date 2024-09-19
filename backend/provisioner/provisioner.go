package provisioner

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"
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
}

var _ provisionerconnect.ProvisionerServiceHandler = (*Service)(nil)

func New(ctx context.Context, config Config, controllerClient ftlv1connect.ControllerServiceClient, devel bool) (*Service, error) {
	return &Service{
		controllerClient: controllerClient,
	}, nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	// TODO: provision infrastructure
	response, err := s.controllerClient.CreateDeployment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call to ftl-controller failed: %w", err)
	}
	return response, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
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
