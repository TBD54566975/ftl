// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	"context"
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
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/unstoppable"
	"github.com/TBD54566975/ftl/observability"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
)

type Config struct {
	Endpoint             *url.URL      `help:"Endpoint the Runner should bind to and advertise." default:"http://localhost:8893"`
	ControlPlaneEndpoint *url.URL      `name:"ftl-endpoint" help:"Control Plane endpoint." env:"FTL_ENDPOINT" default:"http://localhost:8892"`
	DeploymentDir        string        `help:"Directory to store deployments in." default:"${deploymentdir}"`
	Language             string        `help:"Language to advertise for deployments." env:"FTL_LANGUAGE" required:""`
	HeartbeatPeriod      time.Duration `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter      time.Duration `help:"Jitter to add to heartbeat period." default:"2s"`
}

func Start(ctx context.Context, config Config) error {
	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, config.ControlPlaneEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, client)
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL Runner")
	logger.Infof("Deployment directory: %s", config.DeploymentDir)
	err := os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment directory")
	}
	logger.Infof("Using FTL endpoint: %s", config.ControlPlaneEndpoint)
	logger.Infof("Listening on %s", config.Endpoint)

	controlplaneClient := rpc.Dial(ftlv1connect.NewControlPlaneServiceClient, config.ControlPlaneEndpoint.String(), log.Error)

	svc := &Service{
		key:                model.NewRunnerKey(),
		config:             config,
		controlPlaneClient: controlplaneClient,
		forceUpdate:        make(chan struct{}, 8),
	}
	svc.registrationFailure.Store(types.Some(errors.New("not registered with ControlPlane")))

	retry := backoff.Backoff{Max: config.HeartbeatPeriod, Jitter: true}
	go rpc.RetryStreamingClientStream(ctx, retry, svc.controlPlaneClient.RegisterRunner, svc.registrationLoop)

	observabilityClient := rpc.Dial(ftlv1connect.NewObservabilityServiceClient, config.ControlPlaneEndpoint.String(), log.Error)

	obs := observability.NewService(svc.key, observabilityClient)

	return rpc.Serve(ctx, config.Endpoint,
		rpc.Route("/"+ftlv1connect.VerbServiceName+"/", svc), // The Runner proxies all verbs to the deployment.
		rpc.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
		rpc.RawGRPC(v1connect.NewMetricsServiceHandler, obs),
	)
}

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ http.Handler = (*Service)(nil)

type deployment struct {
	key    model.DeploymentKey
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	proxy  *httputil.ReverseProxy
	// Cancelled when plugin terminates
	ctx context.Context
}

type Service struct {
	key model.RunnerKey
	// We use double-checked locking around the atomic so that the read fast-path is lock-free.
	lock       sync.Mutex
	state      atomic.Value[ftlv1.RunnerState]
	deployment atomic.Value[types.Option[*deployment]]

	config              Config
	controlPlaneClient  ftlv1connect.ControlPlaneServiceClient
	registrationFailure atomic.Value[types.Option[error]]
	forceUpdate         chan struct{} // Force an update to be sent to the ControlPlane immediately.
}

// ServeHTTP proxies through to the deployment.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if deployment, ok := s.deployment.Load().Get(); ok {
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

func (s *Service) DeployToRunner(ctx context.Context, req *connect.Request[ftlv1.DeployToRunnerRequest]) (response *connect.Response[ftlv1.DeployToRunnerResponse], err error) {
	if err, ok := s.registrationFailure.Load().Get(); ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to register runner"))
	}

	id, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	// Double-checked lock.
	if s.deployment.Load().Ok() {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("already deployed"))
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("already deployed"))
	}

	s.state.Store(ftlv1.RunnerState_RESERVED)
	defer func() {
		if err != nil {
			s.state.Store(ftlv1.RunnerState_IDLE)
		}
	}()

	gdResp, err := s.controlPlaneClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: req.Msg.DeploymentKey}))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module")
	}
	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, id.String())
	err = os.MkdirAll(deploymentDir, 0700)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deployment directory")
	}
	err = download.Artefacts(ctx, s.controlPlaneClient, id, deploymentDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download artefacts")
	}
	deployment, cmdCtx, err := plugin.Spawn(
		unstoppable.Context(ctx),
		gdResp.Msg.Schema.Name,
		deploymentDir,
		"./main",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars(
			"FTL_ENDPOINT="+s.config.ControlPlaneEndpoint.String(),
			"FTL_OBSERVABILITY_ENDPOINT="+s.config.Endpoint.String(),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to spawn plugin")
	}
	s.deployment.Store(types.Some(s.makePluginProxy(cmdCtx, id, deployment)))
	s.state.Store(ftlv1.RunnerState_ASSIGNED)
	s.forceUpdate <- struct{}{}
	return connect.NewResponse(&ftlv1.DeployToRunnerResponse{}), nil
}

func (s *Service) makePluginProxy(ctx context.Context, key model.DeploymentKey, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *deployment {
	return &deployment{
		ctx:    ctx,
		key:    key,
		plugin: plugin,
		proxy:  httputil.NewSingleHostReverseProxy(plugin.Endpoint),
	}
}

func (s *Service) registrationLoop(ctx context.Context, send func(*ftlv1.RegisterRunnerRequest) error) error {
	logger := log.FromContext(ctx)

	// Figure out the appropriate state.
	state := s.state.Load()
	var errPtr *string
	var deploymenyKey *string
	depl, ok := s.deployment.Load().Get()
	if ok {
		dkey := depl.key.String()
		deploymenyKey = &dkey
		select {
		case <-depl.ctx.Done():
			state = ftlv1.RunnerState_IDLE
			err := context.Cause(depl.ctx)
			errStr := err.Error()
			errPtr = &errStr
			logger.Errorf(err, "Deployment terminated.")
			s.deployment.Store(types.None[*deployment]())

		default:
			state = ftlv1.RunnerState_ASSIGNED
		}
		s.state.Store(state)
	}

	err := send(&ftlv1.RegisterRunnerRequest{
		Key:        s.key.String(),
		Language:   s.config.Language,
		Endpoint:   s.config.Endpoint.String(),
		Deployment: deploymenyKey,
		State:      state,
		Error:      errPtr,
	})
	if err != nil {
		s.registrationFailure.Store(types.Some(err))
		return errors.WithStack(err)
	}
	s.registrationFailure.Store(types.None[error]())
	delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
	logger.Tracef("Registered with ControlPlane, next heartbeat in %s", delay)
	select {

	case <-ctx.Done():
		err = context.Cause(ctx)
		s.registrationFailure.Store(types.Some(err))
		return errors.WithStack(err)

	case <-s.forceUpdate:

	case <-time.After(delay):
	}
	return nil
}
