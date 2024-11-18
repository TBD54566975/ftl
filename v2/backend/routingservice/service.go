package routingservice

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v2alpha1/v2alpha1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

type Config struct {
	SchemaServiceURL    *url.URL             `help:"The URL of the schema service." default:"http://127.0.0.1:9992" env:"FTL_SCHEMA_SERVICE_ENDPOINT"`
	Bind                *url.URL             `help:"The address to bind the service to." default:"http://127.0.0.1:9993"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11n-"`
}

var _ v2alpha1connect.RoutingServiceHandler = (*Service)(nil)

func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx).Scope("routingservice")
	ctx, doneFunc := context.WithCancel(ctx)
	defer doneFunc()
	err := observability.Init(ctx, false, "", "ftl-routing", ftl.Version, config.ObservabilityConfig)
	if err != nil {
		return fmt.Errorf("failed to initialise observability: %w", err)
	}
	schemaService := rpc.Dial(v2alpha1connect.NewSchemaServiceClient, config.SchemaServiceURL.String(), log.Error)
	svc := New(ctx, schemaService)
	logger.Debugf("Starting RoutingService on %s", config.Bind)
	err = rpc.Serve(ctx, config.Bind,
		rpc.GRPC(v2alpha1connect.NewRoutingServiceHandler, svc),
		rpc.HealthCheck(svc.healthCheck),
	)
	if err != nil {
		return fmt.Errorf("SchemaService: %w", err)
	}
	return nil
}

type Service struct {
	routes *xsync.MapOf[string, ftlv1connect.VerbServiceClient]
}

func New(ctx context.Context, client v2alpha1connect.SchemaServiceClient) *Service {
	svc := &Service{
		routes: xsync.NewMapOf[string, ftlv1connect.VerbServiceClient](),
	}
	go svc.updateRoutingTable(ctx, client)
	return svc
}

func (s *Service) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	client, ok := s.routes.Load(req.Msg.Verb.Module)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("%s", req.Msg.Verb.Module))
	}
	resp, err := client.Call(ctx, connect.NewRequest(req.Msg))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", req.Msg.Verb.Module, err)
	}
	return resp, nil
}

func (s *Service) updateRoutingTable(ctx context.Context, client v2alpha1connect.SchemaServiceClient) {
	logger := log.FromContext(ctx).Scope("routingservice")
	rpc.RetryStreamingServerStream(ctx, "routing", backoff.Backoff{Max: time.Second * 2}, &v2alpha1.PullSchemaRequest{}, client.PullSchema,
		func(ctx context.Context, resp *v2alpha1.PullSchemaResponse) error {
			module, err := schema.ModuleFromProto(resp.Schema)
			if err != nil {
				return fmt.Errorf("failed to parse schema: %w", err)
			}

			if resp.Deleted {
				logger.Debugf("Deleting route for %s", module.Name)
				s.routes.Delete(module.Name)
			} else if module.Runtime != nil {
				logger.Debugf("Upserting route for %s: %s", module.Name, module.Runtime.Endpoint)
				if err := s.updateRoute(module); err != nil {
					return fmt.Errorf("failed to update route: %w", err)
				}
			}
			return nil
		},
		rpc.AlwaysRetry(),
	)
}

func (s *Service) updateRoute(module *schema.Module) error {
	endpoint, err := url.Parse(module.Runtime.Endpoint)
	if err != nil {
		return fmt.Errorf("%s: invalid endpoint %q: %w", module.Name, module.Runtime.Endpoint, err)
	}
	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint.String(), log.Error)
	s.routes.Store(module.Name, client)
	return nil
}
