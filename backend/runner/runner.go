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
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/identity"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	ftlobservability "github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/unstoppable"
)

type Config struct {
	Config                []string            `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
	Bind                  *url.URL            `help:"Endpoint the Runner should bind to and advertise." default:"http://127.0.0.1:8893" env:"FTL_RUNNER_BIND"`
	Key                   model.RunnerKey     `help:"Runner key (auto)."`
	ControllerEndpoint    *url.URL            `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	ControllerPublicKey   *identity.PublicKey `name:"ftl-public-key" help:"Controller public key in Base64. Temporarily optional." env:"FTL_CONTROLLER_PUBLIC_KEY"`
	TemplateDir           string              `help:"Template directory to copy into each deployment, if any." type:"existingdir"`
	DeploymentDir         string              `help:"Directory to store deployments in." default:"${deploymentdir}"`
	DeploymentKeepHistory int                 `help:"Number of deployments to keep history for." default:"3"`
	Language              []string            `short:"l" help:"Languages the runner supports." env:"FTL_LANGUAGE" default:"go,kotlin,rust,java"`
	HeartbeatPeriod       time.Duration       `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter       time.Duration       `help:"Jitter to add to heartbeat period." default:"2s"`
	Deployment            string              `help:"The deployment this runner is for." env:"FTL_DEPLOYMENT"`
	DebugPort             int                 `help:"The port to use for debugging." env:"FTL_DEBUG_PORT"`
}

func Start(ctx context.Context, config Config) error {
	ctx, doneFunc := context.WithCancel(ctx)
	defer doneFunc()
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

	// TODO: Retry loop, RetryStreamingClientStreamish
	var identityStore *identity.Store
	if config.ControllerPublicKey != nil {
		identityStore, err = newIdentityStore(ctx, config, key, controllerClient)
		if err != nil {
			observability.Runner.StartupFailed(ctx)
			return fmt.Errorf("failed to create identity store: %w", err)
		}
	}

	svc := &Service{
		key:                key,
		identity:           identityStore,
		config:             config,
		controllerClient:   controllerClient,
		labels:             labels,
		deploymentLogQueue: make(chan log.Entry, 10000),
		cancelFunc:         doneFunc,
	}
	err = svc.deploy(ctx)
	if err != nil {
		// If we fail to deploy we just exit
		// Kube or local scaling will start a new instance to continue
		// This approach means we don't have to handle error states internally
		// It is managed externally by the scaling system
		return err
	}

	go func() {
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.RegisterRunner, svc.registrationLoop)
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, controllerClient.StreamDeploymentLogs, svc.streamLogsLoop)
	}()

	return rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc.HTTP("/", svc),
		rpc.HealthCheck(svc.healthCheck),
	)
}

func newIdentityStore(ctx context.Context, config Config, key model.RunnerKey, controllerClient ftlv1connect.ControllerServiceClient) (*identity.Store, error) {
	controllerVerifier, err := identity.NewVerifier(*config.ControllerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create controller verifier: %w", err)
	}

	identityStore, err := identity.NewStoreNewKeys(identity.NewRunner(key, config.Deployment))
	if err != nil {
		return nil, fmt.Errorf("failed to create identity store: %w", err)
	}

	certRequest, err := identityStore.NewGetCertificateRequest()
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %w", err)
	}

	certResp, err := controllerClient.GetCertification(ctx, connect.NewRequest(&certRequest))
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	certificate, err := identity.NewCertificate(certResp.Msg.Certificate)
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	if err = identityStore.SetCertificate(certificate, controllerVerifier); err != nil {
		observability.Runner.StartupFailed(ctx)
		return nil, fmt.Errorf("failed to set certificate: %w", err)
	}

	return identityStore, nil
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
				// This is not a fatal error, just log it.
				logger.Errorf(err, "Failed to remove old deployment: %s", deployment.Name())
			}
		}
	}

	return nil
}

var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type deployment struct {
	key    model.DeploymentKey
	plugin *plugin.Plugin[ftlv1connect.VerbServiceClient]
	// Cancelled when plugin terminates
	ctx context.Context
}

type Service struct {
	key        model.RunnerKey
	identity   *identity.Store
	lock       sync.Mutex
	deployment atomic.Value[optional.Option[*deployment]]

	config           Config
	controllerClient ftlv1connect.ControllerServiceClient
	// Failed to register with the Controller
	registrationFailure atomic.Value[optional.Option[error]]
	labels              *structpb.Struct
	deploymentLogQueue  chan log.Entry
	cancelFunc          func()
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("no deployment"))
	}
	response, err := deployment.plugin.Client.Call(ctx, req)
	if err != nil {
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(err, "Call to deployments %s failed to perform gRPC call", deployment.key)
		return nil, connect.NewError(connect.CodeOf(err), err)
	} else if response.Msg.GetError() != nil {
		// This is a user level error (i.e. something wrong in the users app)
		// Log it to the deployment logger
		deploymentLogger := s.getDeploymentLogger(ctx, deployment.key)
		deploymentLogger.Errorf(fmt.Errorf("%v", response.Msg.GetError().GetMessage()), "Call to deployments %s failed", deployment.key)
	}

	return connect.NewResponse(response.Msg), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) deploy(ctx context.Context) error {
	logger := log.FromContext(ctx)
	if err, ok := s.registrationFailure.Load().Get(); ok {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to register runner: %w", err))
	}

	key, err := model.ParseDeploymentKey(s.config.Deployment)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.None[string]())
		s.cancelFunc()
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	observability.Deployment.Started(ctx, key.String())
	defer observability.Deployment.Completed(ctx, key.String())

	deploymentLogger := s.getDeploymentLogger(ctx, key)
	ctx = log.ContextWithLogger(ctx, deploymentLogger)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.New("already deployed")
	}

	gdResp, err := s.controllerClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: s.config.Deployment}))
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("invalid module: %w", err)
	}
	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, key.String())
	if s.config.TemplateDir != "" {
		err = copy.Copy(s.config.TemplateDir, deploymentDir)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to copy template directory: %w", err)
		}
	} else {
		err = os.MkdirAll(deploymentDir, 0700)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to create deployment directory: %w", err)
		}
	}
	err = download.Artefacts(ctx, s.controllerClient, key, deploymentDir)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to download artefacts: %w", err)
	}

	envVars := []string{"FTL_ENDPOINT=" + s.config.ControllerEndpoint.String(),
		"FTL_CONFIG=" + strings.Join(s.config.Config, ","),
		"FTL_OBSERVABILITY_ENDPOINT=" + s.config.ControllerEndpoint.String()}
	if s.config.DebugPort > 0 {
		envVars = append(envVars, fmt.Sprintf("FTL_DEBUG_PORT=%d", s.config.DebugPort))
	}

	verbCtx := log.ContextWithLogger(ctx, deploymentLogger.Attrs(map[string]string{"module": module.Name}))
	deployment, cmdCtx, err := plugin.Spawn(
		unstoppable.Context(verbCtx),
		log.FromContext(ctx).GetLevel(),
		gdResp.Msg.Schema.Name,
		deploymentDir,
		"./launch",
		ftlv1connect.NewVerbServiceClient,
		plugin.WithEnvars(
			envVars...,
		),
	)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return fmt.Errorf("failed to spawn plugin: %w", err)
	}

	dep := s.makeDeployment(cmdCtx, key, deployment)
	s.deployment.Store(optional.Some(dep))
	logger.Debugf("Deployed %s", key)
	context.AfterFunc(ctx, func() {
		err := s.Close()
		if err != nil {
			logger := log.FromContext(ctx)
			logger.Errorf(err, "failed to terminate deployment")
		}
	})

	return nil
}

func (s *Service) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	depl, ok := s.deployment.Load().Get()
	if !ok {
		return connect.NewError(connect.CodeNotFound, errors.New("no deployment"))
	}
	// Soft kill.
	err := depl.plugin.Cmd.Kill(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to kill plugin: %w", err)
	}
	// Hard kill after 10 seconds.
	select {
	case <-depl.ctx.Done():
	case <-time.After(10 * time.Second):
		err := depl.plugin.Cmd.Kill(syscall.SIGKILL)
		if err != nil {
			// Should we os.Exit(1) here?
			return fmt.Errorf("failed to kill plugin: %w", err)
		}
	}
	s.deployment.Store(optional.None[*deployment]())
	return nil

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
	var deploymentKey *string
	depl, ok := s.deployment.Load().Get()
	if ok {
		dkey := depl.key.String()
		deploymentKey = &dkey
		select {
		case <-depl.ctx.Done():
			err := context.Cause(depl.ctx)
			s.getDeploymentLogger(ctx, depl.key).Errorf(err, "Deployment terminated")
			s.deployment.Store(optional.None[*deployment]())
			s.cancelFunc()
			return nil
		default:
		}
	}

	logger.Tracef("Registering with Controller for deployment %s", s.config.Deployment)
	err := send(&ftlv1.RegisterRunnerRequest{
		Key:        s.key.String(),
		Endpoint:   s.config.Bind.String(),
		Labels:     s.labels,
		Deployment: s.config.Deployment,
	})
	if err != nil {
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey))
		return fmt.Errorf("failed to register with Controller: %w", err)
	}

	// Wait for the next heartbeat.
	delay := s.config.HeartbeatPeriod + time.Duration(rand.Intn(int(s.config.HeartbeatJitter))) //nolint:gosec
	logger.Tracef("Registered with Controller, next heartbeat in %s", delay)
	observability.Runner.Registered(ctx, optional.Ptr(deploymentKey))
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
		s.registrationFailure.Store(optional.Some(err))
		observability.Runner.RegistrationFailure(ctx, optional.Ptr(deploymentKey))
		return err

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
	ctx = ftlobservability.AddSpanContextToLogger(ctx)

	sink := newDeploymentLogsSink(s.deploymentLogQueue)
	return log.FromContext(ctx).AddSink(sink).Attrs(attrs)
}

func (s *Service) healthCheck(writer http.ResponseWriter, request *http.Request) {
	if s.deployment.Load().Ok() {
		writer.WriteHeader(http.StatusOK)
		return
	}
	writer.WriteHeader(http.StatusServiceUnavailable)
}
