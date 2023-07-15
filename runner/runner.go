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
	"syscall"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/unstoppable"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
)

type Config struct {
	Endpoint             *url.URL        `help:"Endpoint the Runner should bind to and advertise." default:"http://localhost:8893"`
	Key                  model.RunnerKey `help:"Runner key (auto)." placeholder:"R<ULID>" required:"R00000000000000000000000000"`
	ControlPlaneEndpoint *url.URL        `name:"ftl-endpoint" help:"Control Plane endpoint." env:"FTL_ENDPOINT" default:"http://localhost:8892"`
	DeploymentDir        string          `help:"Directory to store deployments in." default:"${deploymentdir}"`
	Language             string          `help:"Language to advertise for deployments." env:"FTL_LANGUAGE" required:""`
	HeartbeatPeriod      time.Duration   `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter      time.Duration   `help:"Jitter to add to heartbeat period." default:"2s"`
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

	key := config.Key
	if key == (model.RunnerKey{}) {
		key = model.NewRunnerKey()
	}

	svc := &Service{
		key:                key,
		config:             config,
		controlPlaneClient: controlplaneClient,
		forceUpdate:        make(chan struct{}, 16),
	}
	svc.state.Store(ftlv1.RunnerState_RUNNER_IDLE)

	go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, controlplaneClient.RegisterRunner, svc.registrationLoop)

	return rpc.Serve(ctx, config.Endpoint,
		rpc.Route("/"+ftlv1connect.VerbServiceName+"/", svc), // The Runner proxies all verbs to the deployment.
		rpc.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
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
	key         model.RunnerKey
	lock        sync.Mutex
	state       atomic.Value[ftlv1.RunnerState]
	forceUpdate chan struct{}
	deployment  atomic.Value[types.Option[*deployment]]

	config             Config
	controlPlaneClient ftlv1connect.ControlPlaneServiceClient
	// Failed to register with the ControlPlane
	registrationFailure atomic.Value[types.Option[error]]
}

func (s *Service) Reserve(ctx context.Context, c *connect.Request[ftlv1.ReserveRequest]) (*connect.Response[ftlv1.ReserveResponse], error) {
	if !s.state.CompareAndSwap(ftlv1.RunnerState_RUNNER_IDLE, ftlv1.RunnerState_RUNNER_RESERVED) {
		return nil, errors.Errorf("can only reserve from IDLE state, not %s", s.state.Load())
	}
	return connect.NewResponse(&ftlv1.ReserveResponse{}), nil
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
	// var notReady *string
	// if err, ok := s.registrationFailure.Load().Get(); ok {
	// 	msg := err.Error()
	// 	notReady = &msg
	// }
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Deploy(ctx context.Context, req *connect.Request[ftlv1.DeployRequest]) (response *connect.Response[ftlv1.DeployResponse], err error) {
	if err, ok := s.registrationFailure.Load().Get(); ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to register runner"))
	}

	id, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("already deployed"))
	}

	// Set the state of the Runner, and update the ControlPlane.
	setState := func(state ftlv1.RunnerState) {
		s.state.Store(state)
		s.forceUpdate <- struct{}{}
	}

	setState(ftlv1.RunnerState_RUNNER_RESERVED)
	defer func() {
		if err != nil {
			setState(ftlv1.RunnerState_RUNNER_IDLE)
			s.deployment.Store(types.None[*deployment]())
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
	setState(ftlv1.RunnerState_RUNNER_ASSIGNED)
	return connect.NewResponse(&ftlv1.DeployResponse{}), nil
}

func (s *Service) Terminate(ctx context.Context, c *connect.Request[ftlv1.TerminateRequest]) (*connect.Response[ftlv1.RunnerHeartbeat], error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	depl, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no deployment"))
	}
	deploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}
	if depl.key != deploymentKey {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("deployment key mismatch"))
	}

	// Soft kill.
	err = depl.plugin.Cmd.Kill(syscall.SIGTERM)
	if err != nil {
		return nil, errors.Wrap(err, "failed to kill plugin")
	}
	// Hard kill after 10 seconds.
	select {
	case <-depl.ctx.Done():
	case <-time.After(10 * time.Second):
		err := depl.plugin.Cmd.Kill(syscall.SIGKILL)
		if err != nil {
			// Should we os.Exit(1) here?
			return nil, errors.Wrap(err, "failed to kill plugin")
		}
	}

	s.deployment.Store(types.None[*deployment]())
	s.state.Store(ftlv1.RunnerState_RUNNER_IDLE)
	return connect.NewResponse(&ftlv1.RunnerHeartbeat{
		Key:      s.key.String(),
		Language: s.config.Language,
		Endpoint: s.config.Endpoint.String(),
		State:    ftlv1.RunnerState_RUNNER_IDLE,
	}), nil
}

func (s *Service) makePluginProxy(ctx context.Context, key model.DeploymentKey, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *deployment {
	return &deployment{
		ctx:    ctx,
		key:    key,
		plugin: plugin,
		proxy:  httputil.NewSingleHostReverseProxy(plugin.Endpoint),
	}
}

func (s *Service) registrationLoop(ctx context.Context, send func(request *ftlv1.RunnerHeartbeat) error) error {
	logger := log.FromContext(ctx)

	// Figure out the appropriate state.
	state := s.state.Load()
	var errPtr *string
	var deploymentKey *string
	depl, ok := s.deployment.Load().Get()
	if ok {
		dkey := depl.key.String()
		deploymentKey = &dkey
		select {
		case <-depl.ctx.Done():
			state = ftlv1.RunnerState_RUNNER_IDLE
			err := context.Cause(depl.ctx)
			errStr := err.Error()
			errPtr = &errStr
			logger.Errorf(err, "Deployment terminated")
			s.deployment.Store(types.None[*deployment]())

		default:
			state = ftlv1.RunnerState_RUNNER_ASSIGNED
		}
		s.state.Store(state)
	}

	logger.Tracef("Registering with ControlPlane as %s", state)
	err := send(&ftlv1.RunnerHeartbeat{
		Key:        s.key.String(),
		Language:   s.config.Language,
		Endpoint:   s.config.Endpoint.String(),
		Deployment: deploymentKey,
		State:      state,
		Error:      errPtr,
	})
	if err != nil {
		logger.Errorf(err, "failed to register with ControlPlane")
		s.registrationFailure.Store(types.Some(err))
		return errors.WithStack(err)
	}
	s.registrationFailure.Store(types.None[error]())

	// Wait for the next heartbeat.
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
