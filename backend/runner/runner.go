// Package runner contains a server that implements the RunnerService and
// proxies VerbService requests to user code.
package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"github.com/otiai10/copy"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/runner/observability"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/unstoppable"
)

type Config struct {
	Config                []string        `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
	Bind                  *url.URL        `help:"Endpoint the Runner should bind to and advertise." default:"http://127.0.0.1:8893" env:"FTL_RUNNER_BIND"`
	Advertise             *url.URL        `help:"Endpoint the Runner should advertise (use --bind if omitted)." default:"" env:"FTL_RUNNER_ADVERTISE"`
	Key                   model.RunnerKey `help:"Runner key (auto)."`
	ControllerEndpoint    *url.URL        `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	TemplateDir           string          `help:"Template directory to copy into each deployment, if any." type:"existingdir"`
	DeploymentDir         string          `help:"Directory to store deployments in." default:"${deploymentdir}"`
	DeploymentKeepHistory int             `help:"Number of deployments to keep history for." default:"3"`
	Language              []string        `short:"l" help:"Languages the runner supports." env:"FTL_LANGUAGE" default:"go,kotlin,rust"`
	HeartbeatPeriod       time.Duration   `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter       time.Duration   `help:"Jitter to add to heartbeat period." default:"2s"`
	RunnerStartDelay      time.Duration   `help:"Time in seconds for a runner to wait before contacting the controller. This can be needed in istio environments to work around initialization races." env:"FTL_RUNNER_START_DELAY" default:"0s"`
}

func Start(ctx context.Context, config Config) error {
	if config.Advertise.String() == "" {
		config.Advertise = config.Bind
	}
	hostname, err := os.Hostname()
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return fmt.Errorf("failed to get hostname: %w", err)
	}
	pid := os.Getpid()

	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, config.ControllerEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, client)

	logger := log.FromContext(ctx).Attrs(map[string]string{"runner": config.Key.String()})
	logger.Debugf("Starting FTL Runner")

	err = manageDeploymentDirectory(logger, config)
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return err
	}

	logger.Debugf("Using FTL endpoint: %s", config.ControllerEndpoint)
	logger.Debugf("Listening on %s", config.Bind)

	controllerClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, config.ControllerEndpoint.String(), log.Error)

	key := config.Key
	if key.IsZero() {
		key = model.NewRunnerKey(config.Bind.Hostname(), config.Bind.Port())
	}
	labels, err := structpb.NewStruct(map[string]any{
		"hostname":  hostname,
		"pid":       pid,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"languages": slices.Map(config.Language, func(t string) any { return t }),
	})
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	svc := &Service{
		key:                key,
		config:             config,
		controllerClient:   controllerClient,
		forceUpdate:        make(chan struct{}, 16),
		labels:             labels,
		deploymentLogQueue: make(chan log.Entry, 10000),
	}
	svc.state.Store(ftlv1.RunnerState_RUNNER_IDLE)

	go func() {
		// In some environments we may want a delay before registering the runner
		// We have seen istio race conditions that we think this will help
		if config.RunnerStartDelay > 0 {
			time.Sleep(config.RunnerStartDelay)
		}
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.RegisterRunner, svc.registrationLoop)
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.StreamDeploymentLogs, svc.streamLogsLoop)
	}()

	return rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc.GRPC(ftlv1connect.NewRunnerServiceHandler, svc),
		rpc.HTTP("/", svc),
	)
}

// manageDeploymentDirectory ensures the deployment directory exists and removes old deployments.
func manageDeploymentDirectory(logger *log.Logger, config Config) error {
	logger.Debugf("Deployment directory: %s", config.DeploymentDir)
	err := os.MkdirAll(config.DeploymentDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Clean up old deployments.
	modules, err := os.ReadDir(config.DeploymentDir)
	if err != nil {
		return fmt.Errorf("failed to read deployment directory: %w", err)
	}

	for _, module := range modules {
		if !module.IsDir() {
			continue
		}

		moduleDir := filepath.Join(config.DeploymentDir, module.Name())
		deployments, err := os.ReadDir(moduleDir)
		if err != nil {
			return fmt.Errorf("failed to read module directory: %w", err)
		}

		if len(deployments) < config.DeploymentKeepHistory {
			continue
		}

		stats, err := slices.MapErr(deployments, func(d os.DirEntry) (os.FileInfo, error) {
			return d.Info()
		})
		if err != nil {
			return fmt.Errorf("failed to stat deployments: %w", err)
		}

		// Sort deployments by modified time, remove anything past the history limit.
		sort.Slice(deployments, func(i, j int) bool {
			return stats[i].ModTime().After(stats[j].ModTime())
		})

		for _, deployment := range deployments[config.DeploymentKeepHistory:] {
			old := filepath.Join(moduleDir, deployment.Name())
			logger.Debugf("Removing old deployment: %s", old)

			err := os.RemoveAll(old)
			if err != nil {
				return fmt.Errorf("failed to remove old deployment: %w", err)
			}
		}
	}

	return nil
}

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type deployment struct {
	key    model.DeploymentKey
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	// Cancelled when plugin terminates
	ctx context.Context
}

type Service struct {
	key         model.RunnerKey
	lock        sync.Mutex
	state       atomic.Value[ftlv1.RunnerState]
	forceUpdate chan struct{}
	deployment  atomic.Value[optional.Option[*deployment]]

	config           Config
	controllerClient ftlv1connect.ControllerServiceClient
	// Failed to register with the Controller
	registrationFailure atomic.Value[optional.Option[error]]
	labels              *structpb.Struct
	deploymentLogQueue  chan log.Entry
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("no deployment"))
	}
	response, err := deployment.plugin.Client.Call(ctx, req)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), err)
	}
	return connect.NewResponse(response.Msg), nil
}

func (s *Service) Reserve(ctx context.Context, c *connect.Request[ftlv1.ReserveRequest]) (*connect.Response[ftlv1.ReserveResponse], error) {
	if !s.state.CompareAndSwap(ftlv1.RunnerState_RUNNER_IDLE, ftlv1.RunnerState_RUNNER_RESERVED) {
		return nil, fmt.Errorf("can only reserve from IDLE state, not %s", s.state.Load())
	}
	return connect.NewResponse(&ftlv1.ReserveResponse{}), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetModuleContext(ctx context.Context, req *connect.Request[ftlv1.ModuleContextRequest], resp *connect.ServerStream[ftlv1.ModuleContextResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("module context must be acquired from the controller"))
}

func (s *Service) AcquireLease(context.Context, *connect.BidiStream[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("leases must be acquired from the controller"))
}

func (s *Service) SendFSMEvent(context.Context, *connect.Request[ftlv1.SendFSMEventRequest]) (*connect.Response[ftlv1.SendFSMEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("FSM events must be sent to the controller"))
}

func (s *Service) PublishEvent(context.Context, *connect.Request[ftlv1.PublishEventRequest]) (*connect.Response[ftlv1.PublishEventResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("topic events must be sent to the controller"))
}

func (s *Service) Deploy(ctx context.Context, req *connect.Request[ftlv1.DeployRequest]) (response *connect.Response[ftlv1.DeployResponse], err error) {
	if err, ok := s.registrationFailure.Load().Get(); ok {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to register runner: %w", err))
	}

	key, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	observability.Deployment.Started(ctx, key.String())
	defer observability.Deployment.Completed(ctx, key.String())

	deploymentLogger := s.getDeploymentLogger(ctx, key)
	ctx = log.ContextWithLogger(ctx, deploymentLogger)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("already deployed"))
	}

	// Set the state of the Runner, and update the Controller.
	setState := func(state ftlv1.RunnerState) {
		s.state.Store(state)
		s.forceUpdate <- struct{}{}
	}

	setState(ftlv1.RunnerState_RUNNER_RESERVED)
	defer func() {
		if err != nil {
			setState(ftlv1.RunnerState_RUNNER_IDLE)
			s.deployment.Store(optional.None[*deployment]())
		}
	}()

	gdResp, err := s.controllerClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: req.Msg.DeploymentKey}))
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, err
	}
	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, fmt.Errorf("invalid module: %w", err)
	}
	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, key.String())
	if s.config.TemplateDir != "" {
		err = copy.Copy(s.config.TemplateDir, deploymentDir)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return nil, fmt.Errorf("failed to copy template directory: %w", err)
		}
	} else {
		err = os.MkdirAll(deploymentDir, 0700)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return nil, fmt.Errorf("failed to create deployment directory: %w", err)
		}
	}
	err = download.Artefacts(ctx, s.controllerClient, key, deploymentDir)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, fmt.Errorf("failed to download artefacts: %w", err)
	}

	verbCtx := log.ContextWithLogger(ctx, deploymentLogger.Attrs(map[string]string{"module": module.Name}))
	deployment, cmdCtx, err := plugin.Spawn(
		unstoppable.Context(verbCtx),
		log.FromContext(ctx).GetLevel(),
		gdResp.Msg.Schema.Name,
		deploymentDir,
		"./main",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars(
			"FTL_ENDPOINT="+s.config.ControllerEndpoint.String(),
			"FTL_CONFIG="+strings.Join(s.config.Config, ","),
			"FTL_OBSERVABILITY_ENDPOINT="+s.config.ControllerEndpoint.String(),
		),
	)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, fmt.Errorf("failed to spawn plugin: %w", err)
	}

	dep := s.makeDeployment(cmdCtx, key, deployment)
	s.deployment.Store(optional.Some(dep))

	setState(ftlv1.RunnerState_RUNNER_ASSIGNED)

	return connect.NewResponse(&ftlv1.DeployResponse{}), nil
}

func (s *Service) Terminate(ctx context.Context, c *connect.Request[ftlv1.TerminateRequest]) (*connect.Response[ftlv1.RegisterRunnerRequest], error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	depl, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no deployment"))
	}
	deploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	if !depl.key.Equal(deploymentKey) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("deployment key mismatch"))
	}

	// Soft kill.
	err = depl.plugin.Cmd.Kill(syscall.SIGTERM)
	if err != nil {
		return nil, fmt.Errorf("failed to kill plugin: %w", err)
	}
	// Hard kill after 10 seconds.
	select {
	case <-depl.ctx.Done():
	case <-time.After(10 * time.Second):
		err := depl.plugin.Cmd.Kill(syscall.SIGKILL)
		if err != nil {
			// Should we os.Exit(1) here?
			return nil, fmt.Errorf("failed to kill plugin: %w", err)
		}
	}
	s.deployment.Store(optional.None[*deployment]())
	s.state.Store(ftlv1.RunnerState_RUNNER_IDLE)
	return connect.NewResponse(&ftlv1.RegisterRunnerRequest{
		Key:      s.key.String(),
		Endpoint: s.config.Advertise.String(),
		State:    ftlv1.RunnerState_RUNNER_IDLE,
		Labels:   s.labels,
	}), nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		http.Error(w, "no deployment", http.StatusNotFound)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(deployment.plugin.Endpoint)
	proxy.ServeHTTP(w, r)

}

func (s *Service) makeDeployment(ctx context.Context, key model.DeploymentKey, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]) *deployment {
	return &deployment{
		ctx:    ctx,
		key:    key,
		plugin: plugin,
	}
}

func (s *Service) registrationLoop(ctx context.Context, send func(request *ftlv1.RegisterRunnerRequest) error) error {
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
			s.getDeploymentLogger(ctx, depl.key).Errorf(err, "Deployment terminated")
			s.deployment.Store(optional.None[*deployment]())

		default:
			state = ftlv1.RunnerState_RUNNER_ASSIGNED
		}
		s.state.Store(state)
	}

	logger.Tracef("Registering with Controller as %s", state)
	err := send(&ftlv1.RegisterRunnerRequest{
		Key:        s.key.String(),
		Endpoint:   s.config.Advertise.String(),
		Labels:     s.labels,
		Deployment: deploymentKey,
		State:      state,
		Error:      errPtr,
	})
	if err != nil {
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey), state)
		return fmt.Errorf("failed to register with Controller: %w", err)
	}
	s.registrationFailure.Store(optional.None[error]())

	// Wait for the next heartbeat.
	delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
	logger.Tracef("Registered with Controller, next heartbeat in %s", delay)
	observability.Runner.Registered(ctx, optional.Ptr(deploymentKey), state)
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey), state)
		return err

	case <-s.forceUpdate:

	case <-time.After(delay):
	}
	return nil
}

func (s *Service) streamLogsLoop(ctx context.Context, send func(request *ftlv1.StreamDeploymentLogsRequest) error) error {
	delay := time.Millisecond * 500

	select {
	case entry := <-s.deploymentLogQueue:
		deploymentKey, ok := entry.Attributes["deployment"]
		if !ok {
			return fmt.Errorf("missing deployment key")
		}

		var errorString *string
		if entry.Error != nil {
			errStr := entry.Error.Error()
			errorString = &errStr
		}
		var request *string
		if reqStr, ok := entry.Attributes["request"]; ok {
			request = &reqStr
		}

		err := send(&ftlv1.StreamDeploymentLogsRequest{
			RequestKey:    request,
			DeploymentKey: deploymentKey,
			TimeStamp:     timestamppb.New(entry.Time),
			LogLevel:      int32(entry.Level.Severity()),
			Attributes:    entry.Attributes,
			Message:       entry.Message,
			Error:         errorString,
		})
		if err != nil {
			return err
		}
	case <-time.After(delay):
	case <-ctx.Done():
		err := context.Cause(ctx)
		return err
	}

	return nil
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentKey model.DeploymentKey) *log.Logger {
	attrs := map[string]string{"deployment": deploymentKey.String()}
	if requestKey, _ := rpc.RequestKeyFromContext(ctx); requestKey.Ok() { //nolint:errcheck // best effort
		attrs["request"] = requestKey.MustGet().String()
	}

	sink := newDeploymentLogsSink(s.deploymentLogQueue)
	return log.FromContext(ctx).AddSink(sink).Attrs(attrs)
}
