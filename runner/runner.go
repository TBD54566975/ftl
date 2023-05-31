// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	context "context"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/google/uuid"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/common/download"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/common/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type Config struct {
	Endpoint        *url.URL      `help:"Endpoint the runner should bind to and advertise." default:"http://localhost:8893"`
	FTLEndpoint     *url.URL      `help:"FTL endpoint." env:"FTL_ENDPOINT" default:"http://localhost:8892"`
	DeploymentDir   string        `help:"Directory to store deployments in." default:"${deploymentdir}"`
	HeartbeatPeriod time.Duration `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter time.Duration `help:"Jitter to add to heartbeat period." default:"2s"`
}

func Start(ctx context.Context, config Config) error {
	ctx = sdkgo.ContextWithClient(ctx, config.FTLEndpoint)
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL runner")
	logger.Infof("Deployment directory: %s", config.DeploymentDir)
	err := os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment directory")
	}
	logger.Infof("Using FTL endpoint: %s", config.FTLEndpoint)
	logger.Infof("Listening on %s", config.Endpoint)

	controlplaneClient := rpc.Dial(ftlv1connect.NewControlPlaneServiceClient, config.FTLEndpoint.String())

	svc := &Service{
		key:                uuid.New(),
		config:             config,
		controlplaneClient: controlplaneClient,
	}
	svc.registrationFailure.Store(types.Some(errors.New("not registered with ControlPlane")))

	go svc.registrationLoop(ctx)

	reflector := grpcreflect.NewStaticReflector(ftlv1connect.RunnerServiceName, ftlv1connect.VerbServiceName)
	return rpc.Serve(ctx, config.Endpoint,
		rpc.Route("/"+ftlv1connect.VerbServiceName+"/", svc), // The Runner proxies all verbs to the deployment.
		rpc.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
		rpc.Route(grpcreflect.NewHandlerV1(reflector)),
		rpc.Route(grpcreflect.NewHandlerV1Alpha(reflector)),
	)
}

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ http.Handler = (*Service)(nil)

type pluginProxy struct {
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	proxy  *httputil.ReverseProxy
}

type Service struct {
	key uuid.UUID
	// We use double-checked locking around the atomic so that the read fast-path is lock-free.
	lock       sync.Mutex
	deployment atomic.Value[*pluginProxy]

	config              Config
	controlplaneClient  ftlv1connect.ControlPlaneServiceClient
	registrationFailure atomic.Value[types.Option[error]]
}

// ServeHTTP proxies through to the deployment.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if deployment := s.deployment.Load(); deployment != nil {
		deployment.proxy.ServeHTTP(w, r)
	} else {
		http.Error(w, "503 No deployment", http.StatusServiceUnavailable)
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	var notReady *string
	if err, ok := s.registrationFailure.Load().Get(); ok {
		msg := err.Error()
		notReady = &msg
	}
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: notReady}), nil
}

func (s *Service) DeployToRunner(ctx context.Context, req *connect.Request[ftlv1.DeployToRunnerRequest]) (*connect.Response[ftlv1.DeployToRunnerResponse], error) {
	if err, ok := s.registrationFailure.Load().Get(); ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to register runner"))
	}

	id, err := uuid.Parse(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	// Double-checked lock.
	if s.deployment.Load() != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("already deployed"))
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load() != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("already deployed"))
	}

	gdResp, err := s.controlplaneClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: req.Msg.DeploymentKey}))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, id.String())
	err = os.Mkdir(deploymentDir, 0700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deployment directory")
	}
	err = download.Artefacts(ctx, s.controlplaneClient, id, s.config.DeploymentDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download artefacts")
	}
	deployment, _, err := plugin.Spawn(
		ctx,
		gdResp.Msg.Schema.Name,
		deploymentDir,
		"./main",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars("FTL_ENDPOINT="+s.config.FTLEndpoint.String()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to spawn plugin")
	}
	s.deployment.Store(s.makePluginProxy(deployment))
	return connect.NewResponse(&ftlv1.DeployToRunnerResponse{}), nil
}

func (s *Service) makePluginProxy(plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *pluginProxy {
	return &pluginProxy{
		plugin: plugin,
		proxy:  httputil.NewSingleHostReverseProxy(plugin.Endpoint),
	}
}

func (s *Service) registrationLoop(ctx context.Context) {
	retry := backoff.Backoff{
		Max:    s.config.HeartbeatPeriod,
		Jitter: true,
	}
	logger := log.FromContext(ctx)
	_ = rpc.RetryStreamingClientStream(ctx, retry, s.controlplaneClient.RegisterRunner,
		func(ctx context.Context, stream *connect.ClientStreamForClient[ftlv1.RegisterRunnerRequest, ftlv1.RegisterRunnerResponse]) (err error) {
			defer func() { s.registrationFailure.Store(types.Some(err)) }()
			for {
				err := stream.Send(&ftlv1.RegisterRunnerRequest{
					Key:      s.key.String(),
					Language: "go",
					Endpoint: s.config.Endpoint.String(),
				})
				if err != nil {
					_, err := stream.CloseAndReceive()
					return errors.WithStack(err)
				}
				s.registrationFailure.Store(types.None[error]())
				delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
				logger.Tracef("Registered with ControlPlane, next heartbeat in %s", delay)
				select {
				case <-ctx.Done():
					return errors.WithStack(context.Cause(ctx))

				case <-time.After(delay):
				}
			}
		},
	)
}
