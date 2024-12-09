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
	mysql "github.com/block/ftl-mysql-auth-proxy"
	"github.com/jpillora/backoff"
	"github.com/otiai10/copy"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	ftldeploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	ftlleaseconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	pubconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1/publishpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/runner/observability"
	"github.com/TBD54566975/ftl/backend/runner/proxy"
	"github.com/TBD54566975/ftl/backend/runner/pubsub"
	"github.com/TBD54566975/ftl/common/plugin"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	ftlobservability "github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/pgproxy"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/unstoppable"
)

type Config struct {
	Config                []string                 `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]" type:"existingfile"`
	Bind                  *url.URL                 `help:"Endpoint the Runner should bind to and advertise." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Key                   model.RunnerKey          `help:"Runner key (auto)."`
	ControllerEndpoint    *url.URL                 `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	LeaseEndpoint         *url.URL                 `name:"ftl-lease-endpoint" help:"Lease endpoint endpoint." env:"FTL_LEASE_ENDPOINT" default:"http://127.0.0.1:8895"`
	TemplateDir           string                   `help:"Template directory to copy into each deployment, if any." type:"existingdir"`
	DeploymentDir         string                   `help:"Directory to store deployments in." default:"${deploymentdir}"`
	DeploymentKeepHistory int                      `help:"Number of deployments to keep history for." default:"3"`
	HeartbeatPeriod       time.Duration            `help:"Minimum period between heartbeats." default:"3s"`
	HeartbeatJitter       time.Duration            `help:"Jitter to add to heartbeat period." default:"2s"`
	Deployment            model.DeploymentKey      `help:"The deployment this runner is for." env:"FTL_DEPLOYMENT"`
	DebugPort             int                      `help:"The port to use for debugging." env:"FTL_DEBUG_PORT"`
	DevEndpoint           optional.Option[url.URL] `help:"An existing endpoint to connect to in development mode" env:"FTL_DEV_ENDPOINT"`
}

func Start(ctx context.Context, config Config, storage *artefacts.OCIArtefactService) error {
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
		"hostname": hostname,
		"pid":      pid,
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
	})
	if err != nil {
		observability.Runner.StartupFailed(ctx)
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	svc := &Service{
		key:                key,
		config:             config,
		storage:            storage,
		controllerClient:   controllerClient,
		labels:             labels,
		deploymentLogQueue: make(chan log.Entry, 10000),
		cancelFunc:         doneFunc,
		devEndpoint:        config.DevEndpoint,
	}

	module, err := svc.getModule(ctx, config.Deployment)
	if err != nil {
		return fmt.Errorf("failed to get module: %w", err)
	}

	startedLatch := &sync.WaitGroup{}
	startedLatch.Add(2)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return svc.startPgProxy(ctx, module, startedLatch)
	})
	g.Go(func() error {
		return svc.startMySQLProxy(ctx, module, startedLatch)
	})
	g.Go(func() error {
		startedLatch.Wait()
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return svc.startDeployment(ctx, config.Deployment, module)
		}
	})

	return fmt.Errorf("failure in runner: %w", g.Wait())
}

func (s *Service) startDeployment(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	err := s.deploy(ctx, key, module)
	if err != nil {
		// If we fail to deploy we just exit
		// Kube or local scaling will start a new instance to continue
		// This approach means we don't have to handle error states internally
		// It is managed externally by the scaling system
		return err
	}
	go func() {
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, s.controllerClient.RegisterRunner, s.registrationLoop)
		go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, s.controllerClient.StreamDeploymentLogs, s.streamLogsLoop)
	}()
	return fmt.Errorf("failure in runner: %w", rpc.Serve(ctx, s.config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s),
		rpc.HTTP("/", s),
		rpc.HealthCheck(s.healthCheck),
	))
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
	key model.DeploymentKey
	// Cancelled when plugin terminates
	ctx      context.Context
	cmd      optional.Option[exec.Cmd]
	endpoint *url.URL // The endpoint the plugin is listening on.
	client   ftlv1connect.VerbServiceClient
}

type Service struct {
	key        model.RunnerKey
	lock       sync.Mutex
	deployment atomic.Value[optional.Option[*deployment]]
	readyTime  atomic.Value[time.Time]

	config           Config
	storage          *artefacts.OCIArtefactService
	controllerClient ftlv1connect.ControllerServiceClient
	// Failed to register with the Controller
	registrationFailure atomic.Value[optional.Option[error]]
	labels              *structpb.Struct
	deploymentLogQueue  chan log.Entry
	cancelFunc          func()
	devEndpoint         optional.Option[url.URL]
	proxy               *proxy.Service
	pubSub              *pubsub.Service
	proxyBindAddress    *url.URL
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	deployment, ok := s.deployment.Load().Get()
	if !ok {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("no deployment"))
	}
	response, err := deployment.client.Call(ctx, req)
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

func (s *Service) getModule(ctx context.Context, key model.DeploymentKey) (*schema.Module, error) {
	gdResp, err := s.controllerClient.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{DeploymentKey: s.config.Deployment.String()}))
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	module, err := schema.ModuleFromProto(gdResp.Msg.Schema)
	if err != nil {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return nil, fmt.Errorf("invalid module: %w", err)
	}
	return module, nil
}

func (s *Service) deploy(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	logger := log.FromContext(ctx)

	if err, ok := s.registrationFailure.Load().Get(); ok {
		observability.Deployment.Failure(ctx, optional.None[string]())
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to register runner: %w", err))
	}

	observability.Deployment.Started(ctx, key.String())
	defer observability.Deployment.Completed(ctx, key.String())

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.deployment.Load().Ok() {
		observability.Deployment.Failure(ctx, optional.Some(key.String()))
		return errors.New("already deployed")
	}

	deploymentLogger := s.getDeploymentLogger(ctx, key)
	ctx = log.ContextWithLogger(ctx, deploymentLogger)

	deploymentDir := filepath.Join(s.config.DeploymentDir, module.Name, key.String())
	if s.config.TemplateDir != "" {
		err := copy.Copy(s.config.TemplateDir, deploymentDir)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to copy template directory: %w", err)
		}
	} else {
		err := os.MkdirAll(deploymentDir, 0700)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to create deployment directory: %w", err)
		}
	}
	var dep *deployment
	if ep, ok := s.devEndpoint.Get(); ok {
		client := rpc.Dial(ftlv1connect.NewVerbServiceClient, ep.String(), log.Error)
		dep = &deployment{
			ctx:      ctx,
			key:      key,
			cmd:      optional.None[exec.Cmd](),
			endpoint: &ep,
			client:   client,
		}
	} else {
		err := download.ArtefactsFromOCI(ctx, s.controllerClient, key, deploymentDir, s.storage)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to download artefacts: %w", err)
		}

		pubSub, err := pubsub.New(module)
		if err != nil {
			observability.Deployment.Failure(ctx, optional.Some(key.String()))
			return fmt.Errorf("failed to create pubsub service: %w", err)
		}
		s.pubSub = pubSub

		deploymentServiceClient := rpc.Dial(ftldeploymentconnect.NewDeploymentServiceClient, s.config.ControllerEndpoint.String(), log.Error)

		ctx = rpc.ContextWithClient(ctx, deploymentServiceClient)

		leaseServiceClient := rpc.Dial(ftlleaseconnect.NewLeaseServiceClient, s.config.LeaseEndpoint.String(), log.Error)

		s.proxy = proxy.New(deploymentServiceClient, leaseServiceClient)

		parse, err := url.Parse("http://127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}
		proxyServer, err := rpc.NewServer(ctx, parse,
			rpc.GRPC(ftlv1connect.NewVerbServiceHandler, s.proxy),
			rpc.GRPC(ftldeploymentconnect.NewDeploymentServiceHandler, s.proxy),
			rpc.GRPC(ftlleaseconnect.NewLeaseServiceHandler, s.proxy),
			rpc.GRPC(pubconnect.NewPublishServiceHandler, s.pubSub),
		)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		urls := proxyServer.Bind.Subscribe(nil)
		go func() {
			err := proxyServer.Serve(ctx)
			if err != nil {
				logger.Errorf(err, "failed to serve")
				return
			}
		}()
		s.proxyBindAddress = <-urls

		logger.Debugf("Setting FTL_ENDPOINT to %s", s.proxyBindAddress.String())
		envVars := []string{"FTL_ENDPOINT=" + s.proxyBindAddress.String(),
			"FTL_CONFIG=" + strings.Join(s.config.Config, ","),
			"FTL_DEPLOYMENT=" + s.config.Deployment.String(),
			"FTL_OBSERVABILITY_ENDPOINT=" + s.config.ControllerEndpoint.String()}
		if s.config.DebugPort > 0 {
			envVars = append(envVars, fmt.Sprintf("FTL_DEBUG_PORT=%d", s.config.DebugPort))
		}

		verbCtx := log.ContextWithLogger(ctx, deploymentLogger.Attrs(map[string]string{"module": module.Name}))
		deployment, cmdCtx, err := plugin.Spawn(
			unstoppable.Context(verbCtx),
			log.FromContext(ctx).GetLevel(),
			module.Name,
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
		dep = s.makeDeployment(cmdCtx, key, deployment)
	}

	s.readyTime.Store(time.Now().Add(time.Second * 2)) // Istio is a bit flakey, add a small delay for readiness
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
	if cmd, ok := depl.cmd.Get(); ok {
		// Soft kill.
		err := cmd.Kill(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("failed to kill plugin: %w", err)
		}
		// Hard kill after 10 seconds.
		select {
		case <-depl.ctx.Done():
		case <-time.After(10 * time.Second):
			err := cmd.Kill(syscall.SIGKILL)
			if err != nil {
				// Should we os.Exit(1) here?
				return fmt.Errorf("failed to kill plugin: %w", err)
			}
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
	proxy := httputil.NewSingleHostReverseProxy(deployment.endpoint)
	proxy.ServeHTTP(w, r)

}

func (s *Service) makeDeployment(ctx context.Context, key model.DeploymentKey, plugin *plugin.Plugin[ftlv1connect.VerbServiceClient, ftlv1.PingRequest, ftlv1.PingResponse, *ftlv1.PingResponse]) *deployment {
	return &deployment{
		ctx:      ctx,
		key:      key,
		cmd:      optional.Ptr(plugin.Cmd),
		endpoint: plugin.Endpoint,
		client:   plugin.Client,
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
		Deployment: s.config.Deployment.String(),
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
		if s.readyTime.Load().After(time.Now()) {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	}
	writer.WriteHeader(http.StatusServiceUnavailable)
}

func (s *Service) startPgProxy(ctx context.Context, module *schema.Module, started *sync.WaitGroup) error {
	logger := log.FromContext(ctx)

	databases := map[string]*schema.Database{}
	for db := range slices.FilterVariants[*schema.Database](module.Decls) {
		if db.Type == "postgres" {
			databases[db.Name] = db
		}
	}

	if len(databases) == 0 {
		started.Done()
		return nil
	}

	// Map the channel to the optional.Option[pgproxy.Started] type.
	channel := make(chan pgproxy.Started)
	go func() {
		select {
		case pgProxy := <-channel:
			os.Setenv("FTL_PROXY_POSTGRES_ADDRESS", fmt.Sprintf("127.0.0.1:%d", pgProxy.Address.Port))
			started.Done()
		case <-ctx.Done():
			started.Done()
			return
		}
	}()

	if err := pgproxy.New("127.0.0.1:0", func(ctx context.Context, params map[string]string) (string, error) {
		db, ok := databases[params["database"]]
		if !ok {
			return "", fmt.Errorf("database %s not found", params["database"])
		}

		dsn, err := dsn.ResolvePostgresDSN(ctx, db.Runtime.Connections.Write)
		if err != nil {
			return "", fmt.Errorf("failed to resolve postgres DSN: %w", err)
		}

		logger.Debugf("Resolved DSN (%s): %s", params["database"], dsn)
		return dsn, nil
	}).Start(ctx, channel); err != nil {
		started.Done()
		return fmt.Errorf("failed to start pgproxy: %w", err)
	}

	return nil
}

func (s *Service) startMySQLProxy(ctx context.Context, module *schema.Module, latch *sync.WaitGroup) error {
	defer latch.Done()
	logger := log.FromContext(ctx)

	databases := map[string]*schema.Database{}
	for _, decl := range module.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Type == "mysql" {
			databases[db.Name] = db
		}
	}

	if len(databases) == 0 {
		return nil
	}
	for db, decl := range databases {
		logger.Debugf("Starting MySQL proxy for %s", db)
		logger := log.FromContext(ctx)
		portC := make(chan int)
		errorC := make(chan error)
		databaseRuntime := decl.Runtime
		var proxy *mysql.Proxy
		switch db := databaseRuntime.Connections.Write.(type) {
		case *schema.DSNDatabaseConnector:
			proxy = mysql.NewProxy("localhost", 0, db.DSN, &mysqlLogger{logger: logger}, portC)
		default:
			return fmt.Errorf("unknown database runtime type: %T", databaseRuntime)
		}
		go func() {
			err := proxy.ListenAndServe(ctx)
			if err != nil {
				errorC <- err
			}
		}()
		port := 0
		select {
		case err := <-errorC:
			return fmt.Errorf("error: %w", err)
		case port = <-portC:
		}

		os.Setenv(strings.ToUpper("FTL_PROXY_MYSQL_ADDRESS_"+decl.Name), fmt.Sprintf("127.0.0.1:%d", port))
	}
	return nil
}

var _ mysql.Logger = (*mysqlLogger)(nil)

type mysqlLogger struct {
	logger *log.Logger
}

func (m *mysqlLogger) Print(v ...any) {
	for _, s := range v {
		m.logger.Infof("mysql: %v", s)
	}
}
