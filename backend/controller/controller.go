package controller

import (
	"bytes"
	"context"
	sha "crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jpillora/backoff"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller/admin"
	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/console"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/dsn"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/ingress"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/leases/dbleaser"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/timeline"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	frontend "github.com/TBD54566975/ftl/frontend/console"
	"github.com/TBD54566975/ftl/internal/configuration"
	cf "github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/cors"
	ftlhttp "github.com/TBD54566975/ftl/internal/http"
	"github.com/TBD54566975/ftl/internal/log"
	ftlmaps "github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	internalobservability "github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
	status "github.com/TBD54566975/ftl/internal/terminal"
)

// CommonConfig between the production controller and development server.
type CommonConfig struct {
	AllowOrigins   []*url.URL    `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_CONTROLLER_ALLOW_ORIGIN"`
	AllowHeaders   []string      `help:"Allow these headers in CORS requests. (Requires AllowOrigins)" env:"FTL_CONTROLLER_ALLOW_HEADERS"`
	NoConsole      bool          `help:"Disable the console."`
	IdleRunners    int           `help:"Number of idle runners to keep around (not supported in production)." default:"3"`
	WaitFor        []string      `help:"Wait for these modules to be deployed before becoming ready." placeholder:"MODULE"`
	CronJobTimeout time.Duration `help:"Timeout for cron jobs." default:"5m"`
}

func (c *CommonConfig) Validate() error {
	if len(c.AllowHeaders) > 0 && len(c.AllowOrigins) == 0 {
		return fmt.Errorf("AllowOrigins must be set when AllowHeaders is used")
	}
	return nil
}

type Config struct {
	Bind                         *url.URL                 `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_CONTROLLER_BIND"`
	IngressBind                  *url.URL                 `help:"Socket to bind to for ingress." default:"http://127.0.0.1:8891" env:"FTL_CONTROLLER_INGRESS_BIND"`
	Key                          model.ControllerKey      `help:"Controller key (auto)." placeholder:"KEY"`
	DSN                          string                   `help:"DAL DSN." default:"${dsn}" env:"FTL_CONTROLLER_DSN"`
	Advertise                    *url.URL                 `help:"Endpoint the Controller should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_CONTROLLER_ADVERTISE"`
	ConsoleURL                   *url.URL                 `help:"The public URL of the console (for CORS)." env:"FTL_CONTROLLER_CONSOLE_URL"`
	ContentTime                  time.Time                `help:"Time to use for console resource timestamps." default:"${timestamp=1970-01-01T00:00:00Z}"`
	RunnerTimeout                time.Duration            `help:"Runner heartbeat timeout." default:"10s"`
	ControllerTimeout            time.Duration            `help:"Controller heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration            `help:"Deployment reservation timeout." default:"120s"`
	ModuleUpdateFrequency        time.Duration            `help:"Frequency to send module updates." default:"30s"`
	EventLogRetention            *time.Duration           `help:"Delete call logs after this time period. 0 to disable" env:"FTL_EVENT_LOG_RETENTION" default:"24h"`
	ArtefactChunkSize            int                      `help:"Size of each chunk streamed to the client." default:"1048576"`
	KMSURI                       *string                  `help:"URI for KMS key e.g. with fake-kms:// or aws-kms://arn:aws:kms:ap-southeast-2:12345:key/0000-1111" env:"FTL_KMS_URI"`
	MaxOpenDBConnections         int                      `help:"Maximum number of database connections." default:"20" env:"FTL_MAX_OPEN_DB_CONNECTIONS"`
	MaxIdleDBConnections         int                      `help:"Maximum number of idle database connections." default:"20" env:"FTL_MAX_IDLE_DB_CONNECTIONS"`
	Registry                     artefacts.RegistryConfig `embed:"" prefix:"oci-"`
	CommonConfig
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c, kong.Vars{"dsn": dsn.DSN("ftl")}); err != nil {
		panic(err)
	}
	if c.Advertise == nil {
		c.Advertise = c.Bind
	}
}

func (c *Config) OpenDBAndInstrument() (*sql.DB, error) {
	conn, err := internalobservability.OpenDBAndInstrument(c.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}
	conn.SetMaxIdleConns(c.MaxIdleDBConnections)
	conn.SetMaxOpenConns(c.MaxOpenDBConnections)
	return conn, nil
}

// Start the Controller. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	runnerScaling scaling.RunnerScaling,
	cm *cf.Manager[configuration.Configuration],
	sm *cf.Manager[configuration.Secrets],
	conn *sql.DB,
	devel bool,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL controller")

	var consoleHandler http.Handler
	var err error
	if config.NoConsole {
		consoleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotImplemented)
			_, _ = w.Write([]byte("Console not installed.")) //nolint:errcheck
		})
	} else {
		consoleHandler, err = frontend.Server(ctx, config.ContentTime, config.Bind, config.ConsoleURL)
		if err != nil {
			return fmt.Errorf("could not start console: %w", err)
		}
		logger.Infof("Web console available at: %s", config.Bind)
	}

	svc, err := New(ctx, conn, cm, sm, config, devel, runnerScaling)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)

	admin := admin.NewAdminService(cm, sm, svc.dal)
	console := console.NewService(svc.dal, svc.timeline)

	ingressHandler := otelhttp.NewHandler(http.Handler(svc), "ftl.ingress")
	if len(config.AllowOrigins) > 0 {
		ingressHandler = cors.Middleware(
			slices.Map(config.AllowOrigins, func(u *url.URL) string { return u.String() }),
			config.AllowHeaders,
			ingressHandler,
		)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		logger.Infof("HTTP ingress server listening on: %s", config.IngressBind)

		return ftlhttp.Serve(ctx, config.IngressBind, ingressHandler)
	})

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewModuleServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewAdminServiceHandler, admin),
			rpc.GRPC(pbconsoleconnect.NewConsoleServiceHandler, console),
			rpc.HTTP("/", consoleHandler),
			rpc.PProf(),
		)
	})
	g.Go(func() error {
		return runnerScaling.Start(ctx, *config.Bind, svc.dbleaser)
	})

	go svc.dal.PollDeployments(ctx)

	return g.Wait()
}

var _ ftlv1connect.ControllerServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type clients struct {
	verb ftlv1connect.VerbServiceClient
}

// ControllerListListener is regularly notified of the current list of controllers
// This is often used to update a hash ring to distribute work.
type ControllerListListener interface {
	UpdatedControllerList(ctx context.Context, controllers []dalmodel.Controller)
}

type Service struct {
	conn               *sql.DB
	dbleaser           *dbleaser.DatabaseLeaser
	dal                *dal.DAL
	key                model.ControllerKey
	deploymentLogsSink *deploymentLogsSink

	cm *cf.Manager[configuration.Configuration]
	sm *cf.Manager[configuration.Secrets]

	tasks                   *scheduledtask.Scheduler
	pubSub                  *pubsub.Service
	registry                artefacts.Service
	timeline                *timeline.Service
	controllerListListeners []ControllerListListener

	// Map from runnerKey.String() to client.
	clients *ttlcache.Cache[string, clients]

	// Complete schema synchronised from the database.
	schemaState    atomic.Value[schemaState]
	schemaSyncLock sync.Mutex

	config Config

	increaseReplicaFailures map[string]int
	asyncCallsLock          sync.Mutex
	runnerScaling           scaling.RunnerScaling

	clientLock sync.Mutex
}

func New(
	ctx context.Context,
	conn *sql.DB,
	cm *cf.Manager[configuration.Configuration],
	sm *cf.Manager[configuration.Secrets],
	config Config,
	devel bool,
	runnerScaling scaling.RunnerScaling,
) (*Service, error) {
	key := config.Key
	if config.Key.IsZero() {
		key = model.NewControllerKey(config.Bind.Hostname(), config.Bind.Port())
	}
	config.SetDefaults()

	// Override some defaults during development mode.
	if devel {
		config.RunnerTimeout = time.Second * 5
		config.ControllerTimeout = time.Second * 5
	}

	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Ptr(config.KMSURI)))
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption dal: %w", err)
	}

	ldb := dbleaser.NewDatabaseLeaser(conn)
	scheduler := scheduledtask.New(ctx, key, ldb)

	svc := &Service{
		cm:                      cm,
		sm:                      sm,
		tasks:                   scheduler,
		dbleaser:                ldb,
		conn:                    conn,
		key:                     key,
		clients:                 ttlcache.New(ttlcache.WithTTL[string, clients](time.Minute)),
		config:                  config,
		increaseReplicaFailures: map[string]int{},
		runnerScaling:           runnerScaling,
	}
	svc.schemaState.Store(schemaState{routes: map[string]Route{}, schema: &schema.Schema{}})

	svc.registry = artefacts.NewOCIRegistryStorage(config.Registry)

	timelineSvc := timeline.New(ctx, conn, encryption)
	svc.timeline = timelineSvc
	pubSub := pubsub.New(ctx, conn, encryption, optional.Some[pubsub.AsyncCallListener](svc), timelineSvc)
	svc.pubSub = pubSub
	svc.dal = dal.New(ctx, conn, encryption, pubSub, svc.registry)

	svc.deploymentLogsSink = newDeploymentLogsSink(ctx, timelineSvc)

	// Use min, max backoff if we are running in production, otherwise use
	// (1s, 1s) (or develBackoff). Will also wrap the job such that it its next
	// runtime is capped at 1s.
	maybeDevelTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) (backoff.Backoff, scheduledtask.Job) {
		if len(develBackoff) > 1 {
			panic("too many devel backoffs")
		}
		chain := job

		// Trace controller operations
		job = func(ctx context.Context) (time.Duration, error) {
			ctx, span := observability.Controller.BeginSpan(ctx, name)
			defer span.End()
			return chain(ctx)
		}

		if devel {
			chain := job
			job = func(ctx context.Context) (time.Duration, error) {
				next, err := chain(ctx)
				// Cap at 1s in development mode.
				return min(next, maxNext), err
			}
			if len(develBackoff) == 1 {
				return develBackoff[0], job
			}
			return backoff.Backoff{Min: time.Second, Max: time.Second}, job
		}
		return makeBackoff(minDelay, maxDelay), job
	}

	parallelTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) {
		maybeDevelJob, backoff := maybeDevelTask(job, name, maxNext, minDelay, maxDelay, develBackoff...)
		svc.tasks.Parallel(name, maybeDevelJob, backoff)
	}

	singletonTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) {
		maybeDevelJob, backoff := maybeDevelTask(job, name, maxNext, minDelay, maxDelay, develBackoff...)
		svc.tasks.Singleton(name, maybeDevelJob, backoff)
	}

	// Parallel tasks.
	parallelTask(svc.syncRoutesAndSchema, "sync-routes-and-schema", time.Second, time.Second, time.Second*5)
	parallelTask(svc.heartbeatController, "controller-heartbeat", time.Second, time.Second*3, time.Second*5)
	parallelTask(svc.updateControllersList, "update-controllers-list", time.Second, time.Second*5, time.Second*5)
	parallelTask(svc.executeAsyncCalls, "execute-async-calls", time.Second, time.Second*5, time.Second*10)

	// This should be a singleton task, but because this is the task that
	// actually expires the leases used to run singleton tasks, it must be
	// parallel.
	parallelTask(svc.expireStaleLeases, "expire-stale-leases", time.Second*2, time.Second, time.Second*5)

	// Singleton tasks use leases to only run on a single controller.
	singletonTask(svc.reapStaleControllers, "reap-stale-controllers", time.Second*2, time.Second*20, time.Second*20)
	singletonTask(svc.reapStaleRunners, "reap-stale-runners", time.Second*2, time.Second, time.Second*10)
	singletonTask(svc.reapCallEvents, "reap-call-events", time.Minute*5, time.Minute, time.Minute*30)
	singletonTask(svc.reapAsyncCalls, "reap-async-calls", time.Second*5, time.Second, time.Second*5)
	return svc, nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	method := strings.ToLower(r.Method)
	requestKey := model.NewRequestKey(model.OriginIngress, fmt.Sprintf("%s %s", method, r.URL.Path))

	sourceAddress := r.RemoteAddr
	err := s.dal.CreateRequest(r.Context(), requestKey, sourceAddress)
	if err != nil {
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), start, optional.Some("failed to create request"))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	routes := s.schemaState.Load().httpRoutes[r.Method]
	if len(routes) == 0 {
		http.NotFound(w, r)
		observability.Ingress.Request(r.Context(), r.Method, r.URL.Path, optional.None[*schemapb.Ref](), start, optional.Some("route not found in dal"))
		return
	}
	ingress.Handle(start, s.schemaState.Load().schema, requestKey, routes, w, r, s.timeline, s.callWithRequest)
}

func (s *Service) ProcessList(ctx context.Context, req *connect.Request[ftlv1.ProcessListRequest]) (*connect.Response[ftlv1.ProcessListResponse], error) {
	processes, err := s.dal.GetProcessList(ctx)
	if err != nil {
		return nil, err
	}
	out, err := slices.MapErr(processes, func(p dal.Process) (*ftlv1.ProcessListResponse_Process, error) {
		var runner *ftlv1.ProcessListResponse_ProcessRunner
		if dbr, ok := p.Runner.Get(); ok {
			labels, err := structpb.NewStruct(dbr.Labels)
			if err != nil {
				return nil, fmt.Errorf("could not marshal labels for runner %s: %w", dbr.Key, err)
			}
			runner = &ftlv1.ProcessListResponse_ProcessRunner{
				Key:      dbr.Key.String(),
				Endpoint: dbr.Endpoint,
				Labels:   labels,
			}
		}
		labels, err := structpb.NewStruct(p.Labels)
		if err != nil {
			return nil, fmt.Errorf("could not marshal labels for deployment %s: %w", p.Deployment, err)
		}
		return &ftlv1.ProcessListResponse_Process{
			Deployment:  p.Deployment.String(),
			MinReplicas: int32(p.MinReplicas),
			Labels:      labels,
			Runner:      runner,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.ProcessListResponse{Processes: out}), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	status, err := s.dal.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get status: %w", err)
	}
	sroutes := s.schemaState.Load().routes
	routes := slices.Map(maps.Values(sroutes), func(route Route) (out *ftlv1.StatusResponse_Route) {
		return &ftlv1.StatusResponse_Route{
			Module:     route.Module,
			Deployment: route.Deployment.String(),
			Endpoint:   route.Endpoint,
		}
	})
	replicas := map[string]int32{}
	protoRunners, err := slices.MapErr(status.Runners, func(r dalmodel.Runner) (*ftlv1.StatusResponse_Runner, error) {
		asString := r.Deployment.String()
		deployment := &asString
		replicas[asString]++
		labels, err := structpb.NewStruct(r.Labels)
		if err != nil {
			return nil, fmt.Errorf("could not marshal attributes for runner %s: %w", r.Key, err)
		}
		return &ftlv1.StatusResponse_Runner{
			Key:        r.Key.String(),
			Endpoint:   r.Endpoint,
			Deployment: deployment,
			Labels:     labels,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	deployments, err := slices.MapErr(status.Deployments, func(d dalmodel.Deployment) (*ftlv1.StatusResponse_Deployment, error) {
		labels, err := structpb.NewStruct(d.Labels)
		if err != nil {
			return nil, fmt.Errorf("could not marshal attributes for deployment %s: %w", d.Key.String(), err)
		}
		return &ftlv1.StatusResponse_Deployment{
			Key:         d.Key.String(),
			Language:    d.Language,
			Name:        d.Module,
			MinReplicas: int32(d.MinReplicas),
			Replicas:    replicas[d.Key.String()],
			Schema:      d.Schema.ToProto().(*schemapb.Module), //nolint:forcetypeassert
			Labels:      labels,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	resp := &ftlv1.StatusResponse{
		Controllers: slices.Map(status.Controllers, func(c dalmodel.Controller) *ftlv1.StatusResponse_Controller {
			return &ftlv1.StatusResponse_Controller{
				Key:      c.Key.String(),
				Endpoint: c.Endpoint,
				Version:  ftl.Version,
			}
		}),
		Runners:     protoRunners,
		Deployments: deployments,
		IngressRoutes: slices.Map(status.IngressRoutes, func(r dalmodel.IngressRouteEntry) *ftlv1.StatusResponse_IngressRoute {
			return &ftlv1.StatusResponse_IngressRoute{
				DeploymentKey: r.Deployment.String(),
				Verb:          &schemapb.Ref{Module: r.Module, Name: r.Verb},
				Method:        r.Method,
				Path:          r.Path,
			}
		}),
		Routes: routes,
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) GetCertification(context.Context, *connect.Request[ftlv1.GetCertificationRequest]) (*connect.Response[ftlv1.GetCertificationResponse], error) {
	panic("implement me")
}

func (s *Service) StreamDeploymentLogs(ctx context.Context, stream *connect.ClientStream[ftlv1.StreamDeploymentLogsRequest]) (*connect.Response[ftlv1.StreamDeploymentLogsResponse], error) {
	for stream.Receive() {
		msg := stream.Msg()
		deploymentKey, err := model.ParseDeploymentKey(msg.DeploymentKey)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
		}
		var requestKey optional.Option[model.RequestKey]
		if msg.RequestKey != nil {
			rkey, err := model.ParseRequestKey(*msg.RequestKey)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request key: %w", err))
			}
			requestKey = optional.Some(rkey)
		}

		s.timeline.EnqueueEvent(ctx, &timeline.Log{
			DeploymentKey: deploymentKey,
			RequestKey:    requestKey,
			Time:          msg.TimeStamp.AsTime(),
			Level:         msg.LogLevel,
			Attributes:    msg.Attributes,
			Message:       msg.Message,
			Error:         optional.Ptr(msg.Error),
		})
	}
	if stream.Err() != nil {
		return nil, stream.Err()
	}
	return connect.NewResponse(&ftlv1.StreamDeploymentLogsResponse{}), nil
}

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	schemas, err := s.dal.GetActiveDeploymentSchemas(ctx)
	if err != nil {
		return nil, err
	}
	modules := []*schemapb.Module{ //nolint:forcetypeassert
		schema.Builtins().ToProto().(*schemapb.Module),
	}
	modules = append(modules, slices.Map(schemas, func(d *schema.Module) *schemapb.Module {
		return d.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	})...)
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: &schemapb.Schema{Modules: modules}}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return s.watchModuleChanges(ctx, func(response *ftlv1.PullSchemaResponse) error {
		return stream.Send(response)
	})
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (response *connect.Response[ftlv1.UpdateDeployResponse], err error) {
	deploymentKey, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	logger := s.getDeploymentLogger(ctx, deploymentKey)
	logger.Debugf("Update deployment for: %s", deploymentKey)

	err = s.dal.SetDeploymentReplicas(ctx, deploymentKey, int(req.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, libdal.ErrNotFound) {
			logger.Errorf(err, "Deployment not found: %s", deploymentKey)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		}
		logger.Errorf(err, "Could not set deployment replicas: %s", deploymentKey)
		return nil, fmt.Errorf("could not set deployment replicas: %w", err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeployResponse{}), nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, c *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	newDeploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	logger := s.getDeploymentLogger(ctx, newDeploymentKey)
	logger.Debugf("Replace deployment for: %s", newDeploymentKey)

	err = s.dal.ReplaceDeployment(ctx, newDeploymentKey, int(c.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, libdal.ErrNotFound) {
			logger.Errorf(err, "Deployment not found: %s", newDeploymentKey)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		} else if errors.Is(err, dal.ErrReplaceDeploymentAlreadyActive) {
			logger.Infof("Reusing deployment: %s", newDeploymentKey)
			dep, err := s.dal.GetDeployment(ctx, newDeploymentKey)
			if err == nil {
				status.UpdateModuleState(ctx, dep.Module, status.BuildStateDeployed)
			} else {
				logger.Errorf(err, "Failed to get deployment from database: %s", newDeploymentKey)
			}
		} else {
			logger.Errorf(err, "Could not replace deployment: %s", newDeploymentKey)
			return nil, fmt.Errorf("could not replace deployment: %w", err)
		}
	}

	return connect.NewResponse(&ftlv1.ReplaceDeployResponse{}), nil
}

func (s *Service) RegisterRunner(ctx context.Context, stream *connect.ClientStream[ftlv1.RegisterRunnerRequest]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {

	deferredDeregistration := false

	logger := log.FromContext(ctx)
	for stream.Receive() {
		msg := stream.Msg()
		endpoint, err := url.Parse(msg.Endpoint)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid endpoint: %w", err))
		}
		if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid endpoint scheme %q", endpoint.Scheme))
		}
		runnerKey, err := model.ParseRunnerKey(msg.Key)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid key: %w", err))
		}

		runnerStr := fmt.Sprintf("%s (%s)", endpoint, runnerKey)
		logger.Tracef("Heartbeat received from runner %s", runnerStr)

		deploymentKey, err := model.ParseDeploymentKey(msg.Deployment)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		err = s.dal.UpsertRunner(ctx, dalmodel.Runner{
			Key:        runnerKey,
			Endpoint:   msg.Endpoint,
			Deployment: deploymentKey,
			Labels:     msg.Labels.AsMap(),
		})
		if errors.Is(err, libdal.ErrConflict) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		} else if err != nil {
			return nil, err
		}
		if !deferredDeregistration {
			// Deregister the runner if the Runner disconnects.
			defer func() {
				err := s.dal.DeregisterRunner(context.Background(), runnerKey)
				if err != nil {
					logger.Errorf(err, "Could not deregister runner %s", runnerStr)
				}
			}()
			deferredDeregistration = true
		}
		_, err = s.syncRoutesAndSchema(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not sync routes: %w", err)
		}
	}
	if stream.Err() != nil {
		return nil, stream.Err()
	}
	return connect.NewResponse(&ftlv1.RegisterRunnerResponse{}), nil
}

func (s *Service) GetDeployment(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return nil, err
	}

	logger := s.getDeploymentLogger(ctx, deployment.Key)
	logger.Debugf("Get deployment for: %s", deployment.Key.String())

	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema:    deployment.Schema.ToProto().(*schemapb.Module), //nolint:forcetypeassert
		Artefacts: slices.Map(deployment.Artefacts, ftlv1.ArtefactToProto),
	}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}

	logger := s.getDeploymentLogger(ctx, deployment.Key)
	logger.Debugf("Get deployment artefacts for: %s", deployment.Key.String())

	chunk := make([]byte, s.config.ArtefactChunkSize)
nextArtefact:
	for _, artefact := range deployment.Artefacts {
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(ftlv1.ArtefactToProto(artefact), clientArtefact) {
				continue nextArtefact
			}
		}
		reader, err := s.registry.Download(ctx, artefact.Digest)
		if err != nil {
			return fmt.Errorf("could not download artefact: %w", err)
		}
		defer reader.Close()
		for {

			n, err := reader.Read(chunk)
			if n != 0 {
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: ftlv1.ArtefactToProto(artefact),
					Chunk:    chunk[:n],
				}); err != nil {
					return fmt.Errorf("could not send artefact chunk: %w", err)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return fmt.Errorf("could not read artefact chunk: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	if len(s.config.WaitFor) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	// It's not actually ready until it is in the routes table
	routes := s.schemaState.Load().routes
	var missing []string
	for _, module := range s.config.WaitFor {
		if _, ok := routes[module]; !ok {
			missing = append(missing, module)
		}
	}
	if len(missing) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	msg := fmt.Sprintf("waiting for deployments: %s", strings.Join(missing, ", "))
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: &msg}), nil
}

// GetModuleContext retrieves config, secrets and DSNs for a module.
func (s *Service) GetModuleContext(ctx context.Context, req *connect.Request[ftlv1.ModuleContextRequest], resp *connect.ServerStream[ftlv1.ModuleContextResponse]) error {
	name := req.Msg.Module

	// Initialize checksum to -1; a zero checksum does occur when the context contains no settings
	lastChecksum := int64(-1)

	for {
		h := sha.New()

		configs, err := s.cm.MapForModule(ctx, name)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get configs: %w", err))
		}
		secrets, err := s.sm.MapForModule(ctx, name)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get secrets: %w", err))
		}
		databases, err := modulecontext.DatabasesFromSecrets(ctx, name, secrets)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get databases: %w", err))
		}

		if err := hashConfigurationMap(h, configs); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on configs: %w", err))
		}
		if err := hashConfigurationMap(h, secrets); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on secrets: %w", err))
		}
		if err := hashDatabaseConfiguration(h, databases); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on databases: %w", err))
		}

		checksum := int64(binary.BigEndian.Uint64((h.Sum(nil))[0:8]))

		if checksum != lastChecksum {
			response := modulecontext.NewBuilder(name).AddConfigs(configs).AddSecrets(secrets).AddDatabases(databases).Build().ToProto()

			if err := resp.Send(response); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("could not send response: %w", err))
			}

			lastChecksum = checksum
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.config.ModuleUpdateFrequency):
		}
	}
}

// hashConfigurationMap computes an order invariant checksum on the configuration
// settings supplied in the map.
func hashConfigurationMap(h hash.Hash, m map[string][]byte) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return fmt.Errorf("error hashing configuration: %w", err)
		}
	}
	return nil
}

// hashDatabaseConfiguration computes an order invariant checksum on the database
// configuration settings supplied in the map.
func hashDatabaseConfiguration(h hash.Hash, m map[string]modulecontext.Database) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), []byte(m[k].DSN)...))
		if err != nil {
			return fmt.Errorf("error hashing database configuration: %w", err)
		}
	}
	return nil
}

// AcquireLease acquires a lease on behalf of a module.
//
// This is a bidirectional stream where each request from the client must be
// responded to with an empty response.
func (s *Service) AcquireLease(ctx context.Context, stream *connect.BidiStream[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]) error {
	var lease leases.Lease
	for {
		msg, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not receive lease request: %w", err))
		}
		if lease == nil {
			lease, _, err = s.dbleaser.AcquireLease(ctx, leases.ModuleKey(msg.Module, msg.Key...), msg.Ttl.AsDuration(), optional.None[any]())
			if err != nil {
				if errors.Is(err, leases.ErrConflict) {
					return connect.NewError(connect.CodeResourceExhausted, fmt.Errorf("lease is held: %w", err))
				}
				return connect.NewError(connect.CodeInternal, fmt.Errorf("could not acquire lease: %w", err))
			}
			defer lease.Release() //nolint:errcheck
		}
		if err = stream.Send(&ftlv1.AcquireLeaseResponse{}); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not send lease response: %w", err))
		}
	}
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	return s.callWithRequest(ctx, headers.CopyRequestForForwarding(req), optional.None[model.RequestKey](), optional.None[model.RequestKey](), "")
}

func (s *Service) PublishEvent(ctx context.Context, req *connect.Request[ftlv1.PublishEventRequest]) (*connect.Response[ftlv1.PublishEventResponse], error) {
	// Publish the event.
	now := time.Now().UTC()
	pubishError := optional.None[string]()
	err := s.pubSub.PublishEventForTopic(ctx, req.Msg.Topic.Module, req.Msg.Topic.Name, req.Msg.Caller, req.Msg.Body)
	if err != nil {
		pubishError = optional.Some(err.Error())
	}

	requestKey := optional.None[string]()
	if rk, err := rpc.RequestKeyFromContext(ctx); err == nil {
		if rk, ok := rk.Get(); ok {
			requestKey = optional.Some(rk.String())
		}
	}

	// Add to timeline.
	sstate := s.schemaState.Load()
	module := req.Msg.Topic.Module
	route, ok := sstate.routes[module]
	if ok {
		s.timeline.EnqueueEvent(ctx, &timeline.PubSubPublish{
			DeploymentKey: route.Deployment,
			RequestKey:    requestKey,
			Time:          now,
			SourceVerb:    schema.Ref{Name: req.Msg.Caller, Module: req.Msg.Topic.Module},
			Topic:         req.Msg.Topic.Name,
			Request:       req.Msg,
			Error:         pubishError,
		})
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to publish a event to topic %s:%s: %w", req.Msg.Topic.Module, req.Msg.Topic.Name, err))
	}
	return connect.NewResponse(&ftlv1.PublishEventResponse{}), nil
}

func (s *Service) callWithRequest(
	ctx context.Context,
	req *connect.Request[ftlv1.CallRequest],
	key optional.Option[model.RequestKey],
	parentKey optional.Option[model.RequestKey],
	sourceAddress string,
) (*connect.Response[ftlv1.CallResponse], error) {
	logger := log.FromContext(ctx)
	start := time.Now()
	ctx, span := observability.Calls.BeginSpan(ctx, req.Msg.Verb)
	defer span.End()

	if req.Msg.Verb == nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: missing verb"))
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("verb is required"))
	}
	if req.Msg.Body == nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: missing body"))
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("body is required"))
	}

	sstate := s.schemaState.Load()
	sch := sstate.schema

	verbRef := schema.RefFromProto(req.Msg.Verb)
	verb := &schema.Verb{}
	logger = logger.Module(verbRef.Module)

	if err := sch.ResolveToType(verbRef, verb); err != nil {
		if errors.Is(err, schema.ErrNotFound) {
			observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb not found"))
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb resolution failed"))
		return nil, err
	}

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get callers"))
		return nil, err
	}

	var currentCaller *schema.Ref // might be nil but that's fine. just means that it's not a cal from another verb
	if len(callers) > 0 {
		currentCaller = callers[len(callers)-1]
	}

	module := verbRef.Module
	route, ok := sstate.routes[module]
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("no routes for module"))
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no routes for module %q", module))
	}

	var requestKey model.RequestKey
	var isNewRequestKey bool
	if k, ok := key.Get(); ok {
		requestKey = k
		isNewRequestKey = false
	} else {
		k, ok, err := headers.GetRequestKey(req.Header())
		if err != nil {
			observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get request key"))
			return nil, err
		} else if !ok {
			requestKey = model.NewRequestKey(model.OriginIngress, "grpc")
			sourceAddress = req.Peer().Addr
			isNewRequestKey = true
		} else {
			requestKey = k
			isNewRequestKey = false
		}
	}
	if isNewRequestKey {
		headers.SetRequestKey(req.Header(), requestKey)
		if err = s.dal.CreateRequest(ctx, requestKey, sourceAddress); err != nil {
			observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to create request"))
			return nil, err
		}
	}

	callEvent := &timeline.Call{
		DeploymentKey:    route.Deployment,
		RequestKey:       requestKey,
		ParentRequestKey: parentKey,
		StartTime:        start,
		DestVerb:         verbRef,
		Callers:          callers,
		Request:          req.Msg,
	}

	if currentCaller != nil && currentCaller.Module != module && !verb.IsExported() {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: verb not exported"))
		err = connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
		callEvent.Response = either.RightOf[*ftlv1.CallResponse](err)
		s.timeline.EnqueueEvent(ctx, callEvent)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
	}

	err = validateCallBody(req.Msg.Body, verb, sch)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid call body"))
		callEvent.Response = either.RightOf[*ftlv1.CallResponse](err)
		s.timeline.EnqueueEvent(ctx, callEvent)
		return nil, err
	}

	client := s.clientsForEndpoint(route.Endpoint)

	if pk, ok := parentKey.Get(); ok {
		ctx = rpc.WithParentRequestKey(ctx, pk)
	}
	ctx = rpc.WithRequestKey(ctx, requestKey)
	ctx = rpc.WithVerbs(ctx, append(callers, verbRef))
	headers.AddCaller(req.Header(), schema.RefFromProto(req.Msg.Verb))

	response, err := client.verb.Call(ctx, req)
	var resp *connect.Response[ftlv1.CallResponse]
	if err == nil {
		resp = connect.NewResponse(response.Msg)
		callEvent.Response = either.LeftOf[error](resp.Msg)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	} else {
		callEvent.Response = either.RightOf[*ftlv1.CallResponse](err)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		logger.Errorf(err, "Call failed to verb %s for deployment %s", verbRef.String(), route.Deployment)
	}

	s.timeline.EnqueueEvent(ctx, callEvent)
	return resp, err
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := slices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, err
	}
	_, need, err := s.registry.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: slices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx)
	digest, err := s.registry.Upload(ctx, artefacts.Artefact{Content: req.Msg.Content})
	if err != nil {
		return nil, err
	}
	logger.Debugf("Created new artefact %s", digest)
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: digest[:]}), nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)

	artefacts := make([]dalmodel.DeploymentArtefact, len(req.Msg.Artefacts))
	for i, artefact := range req.Msg.Artefacts {
		digest, err := sha256.ParseSHA256(artefact.Digest)
		if err != nil {
			logger.Errorf(err, "Invalid digest %s", artefact.Digest)
			return nil, fmt.Errorf("invalid digest: %w", err)
		}
		artefacts[i] = dalmodel.DeploymentArtefact{
			Executable: artefact.Executable,
			Path:       artefact.Path,
			Digest:     digest,
		}
	}
	ms := req.Msg.Schema
	if ms.Runtime == nil {
		err := errors.New("missing runtime metadata")
		logger.Errorf(err, "Missing runtime metadata")
		return nil, err
	}

	module, err := schema.ModuleFromProto(ms)
	if err != nil {
		logger.Errorf(err, "Invalid module schema")
		return nil, fmt.Errorf("invalid module schema: %w", err)
	}
	module, err = s.validateModuleSchema(ctx, module)
	if err != nil {
		logger.Errorf(err, "Invalid module schema")
		return nil, fmt.Errorf("invalid module schema: %w", err)
	}

	for _, d := range module.Decls {
		if db, ok := d.(*schema.Database); ok && db.Runtime != nil {
			key := dsnSecretKey(module.Name, db.Name)

			if err := s.sm.Set(ctx, configuration.NewRef(module.Name, key), db.Runtime.DSN); err != nil {
				return nil, fmt.Errorf("could not set database secret %s: %w", key, err)
			}
			logger.Infof("Database declaration: %s -> %s", db.Name, db.Runtime.DSN)
		}
	}

	ingressRoutes := extractIngressRoutingEntries(req.Msg)

	dkey, err := s.dal.CreateDeployment(ctx, ms.Runtime.Language, module, artefacts, ingressRoutes)
	if err != nil {
		logger.Errorf(err, "Could not create deployment")
		return nil, fmt.Errorf("could not create deployment: %w", err)
	}

	deploymentLogger := s.getDeploymentLogger(ctx, dkey)
	deploymentLogger.Debugf("Created deployment %s", dkey)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: dkey.String()}), nil
}

func (s *Service) ResetSubscription(ctx context.Context, req *connect.Request[ftlv1.ResetSubscriptionRequest]) (*connect.Response[ftlv1.ResetSubscriptionResponse], error) {
	err := s.pubSub.ResetSubscription(ctx, req.Msg.Subscription.Module, req.Msg.Subscription.Name)
	if err != nil {
		return nil, fmt.Errorf("could not reset subscription: %w", err)
	}
	return connect.NewResponse(&ftlv1.ResetSubscriptionResponse{}), nil
}

func dsnSecretKey(module string, db string) string {
	return fmt.Sprintf("FTL_DSN_%s_%s",
		strings.ToUpper(stripNonAlphanumeric(module)),
		strings.ToUpper(stripNonAlphanumeric(db)),
	)
}

func stripNonAlphanumeric(s string) string {
	var result strings.Builder
	for _, r := range s {
		if ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Load schemas for existing modules, combine with our new one, and validate the new module in the context
// of the whole schema.
func (s *Service) validateModuleSchema(ctx context.Context, module *schema.Module) (*schema.Module, error) {
	existingModules, err := s.dal.GetActiveDeploymentSchemas(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get existing schemas: %w", err)
	}
	schemaMap := ftlmaps.FromSlice(existingModules, func(el *schema.Module) (string, *schema.Module) { return el.Name, el })
	schemaMap[module.Name] = module
	fullSchema := &schema.Schema{Modules: maps.Values(schemaMap)}
	schema, err := schema.ValidateModuleInSchema(fullSchema, optional.Some[*schema.Module](module))
	if err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}
	return schema.Module(module.Name).MustGet(), nil
}

func (s *Service) getDeployment(ctx context.Context, key string) (*model.Deployment, error) {
	dkey, err := model.ParseDeploymentKey(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	deployment, err := s.dal.GetDeployment(ctx, dkey)
	if errors.Is(err, pgx.ErrNoRows) {
		logger := s.getDeploymentLogger(ctx, dkey)
		logger.Errorf(err, "Deployment not found")
		return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not retrieve deployment: %w", err))
	}
	return deployment, nil
}

// Return or create the RunnerService and VerbService clients for an endpoint.
func (s *Service) clientsForEndpoint(endpoint string) clients {
	clientItem := s.clients.Get(endpoint)
	if clientItem != nil {
		return clientItem.Value()
	}
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	// Double check it was not added while we were waiting for the lock
	clientItem = s.clients.Get(endpoint)
	if clientItem != nil {
		return clientItem.Value()
	}

	client := clients{
		verb: rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Error),
	}
	s.clients.Set(endpoint, client, time.Minute)
	return client
}

func (s *Service) reapStaleRunners(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)
	count, err := s.dal.KillStaleRunners(context.Background(), s.config.RunnerTimeout)
	if err != nil {
		return 0, fmt.Errorf("failed to delete stale runners: %w", err)
	} else if count > 0 {
		logger.Debugf("Reaped %d stale runners", count)
	}
	return s.config.RunnerTimeout, nil
}

// AsyncCallWasAdded is an optional notification that an async call was added by this controller
//
// It allows us to speed up execution of scheduled async calls rather than waiting for the next poll time.
func (s *Service) AsyncCallWasAdded(ctx context.Context) {
	go func() {
		if _, err := s.executeAsyncCalls(ctx); err != nil {
			log.FromContext(ctx).Errorf(err, "failed to progress subscriptions")
		}
	}()
}

func (s *Service) executeAsyncCalls(ctx context.Context) (interval time.Duration, returnErr error) {
	// There are multiple entry points into this function, but we want the controller to handle async calls one at a time.
	s.asyncCallsLock.Lock()
	defer s.asyncCallsLock.Unlock()

	logger := log.FromContext(ctx)
	logger.Tracef("Acquiring async call")

	now := time.Now().UTC()
	sstate := s.schemaState.Load()

	enqueueTimelineEvent := func(call *dal.AsyncCall, err optional.Option[error]) {
		module := call.Verb.Module
		route, ok := sstate.routes[module]
		if ok {
			eventType := timeline.AsyncExecuteEventTypeUnkown
			switch call.Origin.(type) {
			case async.AsyncOriginCron:
				eventType = timeline.AsyncExecuteEventTypeCron
			case async.AsyncOriginPubSub:
				eventType = timeline.AsyncExecuteEventTypePubSub
			case *async.AsyncOriginPubSub:
				eventType = timeline.AsyncExecuteEventTypePubSub
			default:
				break
			}
			errStr := optional.None[string]()
			if e, ok := err.Get(); ok {
				errStr = optional.Some(e.Error())
			}
			s.timeline.EnqueueEvent(ctx, &timeline.AsyncExecute{
				DeploymentKey: route.Deployment,
				RequestKey:    call.ParentRequestKey,
				EventType:     eventType,
				Verb:          *call.Verb.ToRef(),
				Time:          now,
				Error:         errStr,
			})
		}
	}

	call, leaseCtx, err := s.dal.AcquireAsyncCall(ctx)
	if errors.Is(err, libdal.ErrNotFound) {
		logger.Tracef("No async calls to execute")
		return time.Second * 2, nil
	} else if err != nil {
		if call == nil {
			observability.AsyncCalls.AcquireFailed(ctx, err)
		} else {
			observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, err)
			enqueueTimelineEvent(call, optional.Some(err))
		}
		return 0, err
	}
	// use originalCtx for things that should are done outside of the lease lifespan
	originalCtx := ctx
	ctx = leaseCtx

	// Extract the otel context from the call
	ctx, err = observability.ExtractTraceContextToContext(ctx, call.TraceContext)
	if err != nil {
		observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, err)
		enqueueTimelineEvent(call, optional.Some(err))
		return 0, fmt.Errorf("failed to extract trace context: %w", err)
	}

	// Extract the request key from the call and attach it as the parent request key
	parentRequestKey := optional.None[model.RequestKey]()
	if prk, ok := call.ParentRequestKey.Get(); ok {
		if rk, err := model.ParseRequestKey(prk); err == nil {
			parentRequestKey = optional.Some(rk)
		} else {
			logger.Tracef("Ignoring invalid request key: %s", prk)
		}
	}

	observability.AsyncCalls.Acquired(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, call.Catching, nil)

	defer func() {
		if returnErr == nil {
			// Post-commit notification based on origin
			switch origin := call.Origin.(type) {
			case async.AsyncOriginCron:
				break

			case async.AsyncOriginPubSub:
				go s.pubSub.AsyncCallDidCommit(originalCtx, origin)

			default:
				break
			}
		}

		call.Release() //nolint:errcheck
	}()

	logger = logger.Scope(fmt.Sprintf("%s:%s", call.Origin, call.Verb)).Module(call.Verb.Module)

	if call.Catching {
		// Retries have been exhausted but catch verb has previously failed
		// We need to try again to catch the async call
		return 0, s.catchAsyncCall(ctx, logger, call)
	}

	logger.Tracef("Executing async call")
	req := &ftlv1.CallRequest{
		Verb:     call.Verb.ToProto(),
		Body:     call.Request,
		Metadata: metadataForAsyncCall(call),
	}
	resp, err := s.callWithRequest(ctx, connect.NewRequest(req), optional.None[model.RequestKey](), parentRequestKey, s.config.Advertise.String())
	var callResult either.Either[[]byte, string]
	if err != nil {
		logger.Warnf("Async call could not be called: %v", err)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.Some("async call could not be called"))
		callResult = either.RightOf[[]byte](err.Error())
	} else if perr := resp.Msg.GetError(); perr != nil {
		logger.Warnf("Async call failed: %s", perr.Message)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.Some("async call failed"))
		callResult = either.RightOf[[]byte](perr.Message)
	} else {
		logger.Debugf("Async call succeeded")
		callResult = either.LeftOf[string](resp.Msg.GetBody())
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, optional.None[string]())
	}

	queueDepth := call.QueueDepth
	didScheduleAnotherCall, err := s.dal.CompleteAsyncCall(ctx, call, callResult, func(tx *dal.DAL, isFinalResult bool) error {
		return s.finaliseAsyncCall(ctx, tx, call, callResult, isFinalResult)
	})
	if err != nil {
		observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, queueDepth, err)
		enqueueTimelineEvent(call, optional.Some(err))
		return 0, fmt.Errorf("failed to complete async call: %w", err)
	}
	if !didScheduleAnotherCall {
		// Queue depth is queried at acquisition time, which means it includes the async
		// call that was just executed so we need to decrement
		queueDepth = call.QueueDepth - 1
	}
	observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, false, queueDepth, nil)
	enqueueTimelineEvent(call, optional.None[error]())
	return 0, nil
}

func (s *Service) catchAsyncCall(ctx context.Context, logger *log.Logger, call *dal.AsyncCall) error {
	catchVerb, ok := call.CatchVerb.Get()
	if !ok {
		logger.Warnf("Async call %s could not catch, missing catch verb", call.Verb)
		return fmt.Errorf("async call %s could not catch, missing catch verb", call.Verb)
	}
	logger.Debugf("Catching async call %s with %s", call.Verb, catchVerb)

	sch := s.schemaState.Load().schema

	verb := &schema.Verb{}
	if err := sch.ResolveToType(call.Verb.ToRef(), verb); err != nil {
		logger.Warnf("Async call %s could not catch, could not resolve original verb: %s", call.Verb, err)
		return fmt.Errorf("async call %s could not catch, could not resolve original verb: %w", call.Verb, err)
	}

	originalError := call.Error.Default("unknown error")
	originalResult := either.RightOf[[]byte](originalError)

	request := map[string]any{
		"verb": map[string]string{
			"module": call.Verb.Module,
			"name":   call.Verb.Name,
		},
		"requestType": verb.Request.String(),
		"request":     json.RawMessage(call.Request),
		"error":       originalError,
	}
	body, err := json.Marshal(request)
	if err != nil {
		logger.Warnf("Async call %s could not marshal body while catching", call.Verb)
		return fmt.Errorf("async call %s could not marshal body while catching", call.Verb)
	}

	req := &ftlv1.CallRequest{
		Verb:     catchVerb.ToProto(),
		Body:     body,
		Metadata: metadataForAsyncCall(call),
	}
	resp, err := s.callWithRequest(ctx, connect.NewRequest(req), optional.None[model.RequestKey](), optional.None[model.RequestKey](), s.config.Advertise.String())
	var catchResult either.Either[[]byte, string]
	if err != nil {
		// Could not call catch verb
		logger.Warnf("Async call %s could not call catch verb %s: %s", call.Verb, catchVerb, err)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call could not be called"))
		catchResult = either.RightOf[[]byte](err.Error())
	} else if perr := resp.Msg.GetError(); perr != nil {
		// Catch verb failed
		logger.Warnf("Async call %s had an error while catching (%s): %s", call.Verb, catchVerb, perr.Message)
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call failed"))
		catchResult = either.RightOf[[]byte](perr.Message)
	} else {
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.None[string]())
		catchResult = either.LeftOf[string](resp.Msg.GetBody())
	}
	queueDepth := call.QueueDepth
	didScheduleAnotherCall, err := s.dal.CompleteAsyncCall(ctx, call, catchResult, func(tx *dal.DAL, isFinalResult bool) error {
		// Exposes the original error to external components such as PubSub
		return s.finaliseAsyncCall(ctx, tx, call, originalResult, isFinalResult)
	})
	if err != nil {
		logger.Errorf(err, "Async call %s could not complete after catching (%s)", call.Verb, catchVerb)
		observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, queueDepth, err)
		return fmt.Errorf("async call %s could not complete after catching (%s): %w", call.Verb, catchVerb, err)
	}
	if !didScheduleAnotherCall {
		// Queue depth is queried at acquisition time, which means it includes the async
		// call that was just executed so we need to decrement
		queueDepth = call.QueueDepth - 1
	}
	logger.Debugf("Caught async call %s with %s", call.Verb, catchVerb)
	observability.AsyncCalls.Completed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, queueDepth, nil)
	return nil
}

// fails async calls that have had their leases reaped
func (s *Service) reapAsyncCalls(ctx context.Context) (nextInterval time.Duration, err error) {
	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, fmt.Errorf("could not start transaction: %w", err))
	}
	defer tx.CommitOrRollback(ctx, &err)

	limit := 20
	calls, err := tx.GetZombieAsyncCalls(ctx, 20)
	if err != nil {
		return 0, fmt.Errorf("failed to get zombie async calls: %w", err)
	}
	for _, call := range calls {
		callResult := either.RightOf[[]byte]("async call lease expired")
		_, err := tx.CompleteAsyncCall(ctx, call, callResult, func(tx *dal.DAL, isFinalResult bool) error {
			return s.finaliseAsyncCall(ctx, tx, call, callResult, isFinalResult)
		})
		if err != nil {
			return 0, fmt.Errorf("failed to complete zombie async call: %w", err)
		}
		observability.AsyncCalls.Executed(ctx, call.Verb, call.CatchVerb, call.Origin.String(), call.ScheduledAt, true, optional.Some("async call lease failed"))
	}

	if len(calls) == limit {
		return 0, nil
	}
	return time.Second * 5, nil
}

func metadataForAsyncCall(call *dal.AsyncCall) *ftlv1.Metadata {
	switch call.Origin.(type) {
	case async.AsyncOriginCron:
		return &ftlv1.Metadata{}

	case async.AsyncOriginPubSub:
		return &ftlv1.Metadata{}

	default:
		panic(fmt.Errorf("unsupported async call origin: %v", call.Origin))
	}
}

func (s *Service) finaliseAsyncCall(ctx context.Context, tx *dal.DAL, call *dal.AsyncCall, callResult either.Either[[]byte, string], isFinalResult bool) error {
	_, failed := callResult.(either.Right[[]byte, string])

	// Allow for handling of completion based on origin
	switch origin := call.Origin.(type) {
	case async.AsyncOriginPubSub:
		if err := s.pubSub.OnCallCompletion(ctx, tx.Connection, origin, failed, isFinalResult); err != nil {
			return fmt.Errorf("failed to finalize pubsub async call: %w", err)
		}

	default:
		panic(fmt.Errorf("unsupported async call origin: %v", call.Origin))
	}
	return nil
}

func (s *Service) expireStaleLeases(ctx context.Context) (time.Duration, error) {
	err := s.dbleaser.ExpireLeases(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to expire leases: %w", err)
	}
	return time.Second * 1, nil
}

// Periodically remove stale (ie. have not heartbeat recently) controllers from the database.
func (s *Service) reapStaleControllers(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)
	count, err := s.dal.KillStaleControllers(context.Background(), s.config.ControllerTimeout)
	if err != nil {
		return 0, fmt.Errorf("failed to delete stale controllers: %w", err)
	} else if count > 0 {
		logger.Debugf("Reaped %d stale controllers", count)
	}
	return time.Second * 10, nil
}

// Periodically update the DB with the current state of the controller.
func (s *Service) heartbeatController(ctx context.Context) (time.Duration, error) {
	_, err := s.dal.UpsertController(ctx, s.key, s.config.Advertise.String())
	if err != nil {
		return 0, fmt.Errorf("failed to heartbeat controller: %w", err)
	}
	return time.Second * 3, nil
}

func (s *Service) updateControllersList(ctx context.Context) (time.Duration, error) {
	controllers, err := s.dal.GetActiveControllers(ctx)
	if err != nil {
		return 0, err
	}
	for _, listener := range s.controllerListListeners {
		listener.UpdatedControllerList(ctx, controllers)
	}
	return time.Second * 5, nil
}

func (s *Service) watchModuleChanges(ctx context.Context, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx)
	type moduleStateEntry struct {
		hash        []byte
		minReplicas int
	}
	moduleState := map[string]moduleStateEntry{}
	moduleByDeploymentKey := map[string]string{}

	// Seed the notification channel with the current deployments.
	seedDeployments, err := s.dal.GetActiveDeployments(ctx)
	if err != nil {
		return err
	}
	initialCount := len(seedDeployments)
	deploymentChanges := make(chan dal.DeploymentNotification, 32)
	for _, deployment := range seedDeployments {
		deploymentChanges <- dal.DeploymentNotification{Message: optional.Some(deployment)}
	}
	logger.Debugf("Seeded %d deployments", initialCount)

	builtins := schema.Builtins().ToProto().(*schemapb.Module) //nolint:forcetypeassert
	buildinsResponse := &ftlv1.PullSchemaResponse{
		ModuleName: builtins.Name,
		Schema:     builtins,
		ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
		More:       initialCount > 0,
	}

	err = sendChange(buildinsResponse)
	if err != nil {
		return err
	}

	// Subscribe to deployment changes.
	s.dal.DeploymentChanges.Subscribe(deploymentChanges)
	defer s.dal.DeploymentChanges.Unsubscribe(deploymentChanges)

	for {
		select {
		case <-ctx.Done():
			return nil

		case notification := <-deploymentChanges:
			var response *ftlv1.PullSchemaResponse
			// Deleted key
			if deletion, ok := notification.Deleted.Get(); ok {
				name := moduleByDeploymentKey[deletion.String()]
				response = &ftlv1.PullSchemaResponse{
					ModuleName:    name,
					DeploymentKey: deletion.String(),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED,
				}
				delete(moduleState, name)
				delete(moduleByDeploymentKey, deletion.String())
			} else if message, ok := notification.Message.Get(); ok {
				moduleSchema := message.Schema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
				if moduleSchema.Runtime == nil {
					moduleSchema.Runtime = &schemapb.ModuleRuntime{
						Language: message.Language,
					}
				}
				moduleSchema.Runtime.CreateTime = timestamppb.New(message.CreatedAt)
				moduleSchema.Runtime.MinReplicas = int32(message.MinReplicas)

				hasher := sha.New()
				data := []byte(moduleSchema.String())
				if _, err := hasher.Write(data); err != nil {
					return err
				}

				newState := moduleStateEntry{
					hash:        hasher.Sum(nil),
					minReplicas: message.MinReplicas,
				}
				if current, ok := moduleState[message.Schema.Name]; ok {
					if !bytes.Equal(current.hash, newState.hash) || current.minReplicas != newState.minReplicas {
						changeType := ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED
						// A deployment is considered removed if its minReplicas is set to 0.
						if current.minReplicas > 0 && message.MinReplicas == 0 {
							changeType = ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED
						}
						response = &ftlv1.PullSchemaResponse{
							ModuleName:    moduleSchema.Name,
							DeploymentKey: message.Key.String(),
							Schema:        moduleSchema,
							ChangeType:    changeType,
						}
					}
				} else {
					response = &ftlv1.PullSchemaResponse{
						ModuleName:    moduleSchema.Name,
						DeploymentKey: message.Key.String(),
						Schema:        moduleSchema,
						ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
						More:          initialCount > 1,
					}
					if initialCount > 0 {
						initialCount--
					}
				}
				moduleState[message.Schema.Name] = newState
				delete(moduleByDeploymentKey, message.Key.String()) // The deployment may have changed.
				moduleByDeploymentKey[message.Key.String()] = message.Schema.Name
			}

			if response != nil {
				logger.Tracef("Sending change %s", response.ChangeType)
				err := sendChange(response)
				if err != nil {
					return err
				}
			} else {
				logger.Tracef("No change")
			}
		}
	}
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentKey model.DeploymentKey) *log.Logger {
	attrs := map[string]string{"deployment": deploymentKey.String()}
	if requestKey, _ := rpc.RequestKeyFromContext(ctx); requestKey.Ok() { //nolint:errcheck // best effort?
		attrs["request"] = requestKey.MustGet().String()
	}

	return log.FromContext(ctx).AddSink(s.deploymentLogsSink).Attrs(attrs)
}

// Periodically sync the routing table and schema from the DB.
// We do this in a single function so the routing table and schema are always consistent
// And they both need the same info from the DB
func (s *Service) syncRoutesAndSchema(ctx context.Context) (ret time.Duration, err error) {
	s.schemaSyncLock.Lock() // This can result in confusing log messages if it is called concurrently
	defer s.schemaSyncLock.Unlock()
	deployments, err := s.dal.GetActiveDeployments(ctx)
	if errors.Is(err, libdal.ErrNotFound) {
		deployments = []dalmodel.Deployment{}
	} else if err != nil {
		return 0, err
	}
	tx, err := s.dal.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	old := s.schemaState.Load().routes
	newRoutes := map[string]Route{}
	modulesByName := map[string]*schema.Module{}

	builtins := schema.Builtins().ToProto().(*schemapb.Module) //nolint:forcetypeassert
	modulesByName[builtins.Name], err = schema.ModuleFromProto(builtins)
	if err != nil {
		return 0, fmt.Errorf("failed to convert builtins to schema: %w", err)
	}
	for _, v := range deployments {
		deploymentLogger := s.getDeploymentLogger(ctx, v.Key)
		deploymentLogger.Tracef("processing deployment %s for route table", v.Key.String())
		// Deployments are in order, oldest to newest
		// If we see a newer one overwrite an old one that means the new one is read
		// And we set its replicas to zero
		// It may seem a bit odd to do this here but this is where we are actually updating the routing table
		// Which is what makes as a deployment 'live' from a clients POV
		optURI, err := s.runnerScaling.GetEndpointForDeployment(ctx, v.Module, v.Key.String())
		if err != nil {
			deploymentLogger.Debugf("Failed to get updated endpoint for deployment %s", v.Key.String())
			continue
		} else if uri, ok := optURI.Get(); ok {
			// Check if this is a new route
			targetEndpoint := uri.String()
			if oldRoute, oldRouteExists := old[v.Module]; !oldRouteExists || oldRoute.Deployment.String() != v.Key.String() {
				// If it is a new route we only add it if we can ping it
				// Kube deployments can take a while to come up, so we don't want to add them to the routing table until they are ready.
				_, err := s.clientsForEndpoint(targetEndpoint).verb.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
				if err != nil {
					deploymentLogger.Tracef("Unable to ping %s, not adding to route table", v.Key.String())
					continue
				}
				deploymentLogger.Infof("Deployed %s", v.Key.String())
				status.UpdateModuleState(ctx, v.Module, status.BuildStateDeployed)
			}
			if prev, ok := newRoutes[v.Module]; ok {
				// We have already seen a route for this module, the existing route must be an old one
				// as the deployments are in order
				// We have a new route ready to go, so we can just set the old one to 0 replicas
				// Do this in a TX so it doesn't happen until the route table is updated
				deploymentLogger.Debugf("Setting %s to zero replicas", prev.Deployment)
				err := tx.SetDeploymentReplicas(ctx, prev.Deployment, 0)
				if err != nil {
					deploymentLogger.Errorf(err, "Failed to set replicas to 0 for deployment %s", prev.Deployment.String())
				}
			}
			newRoutes[v.Module] = Route{Module: v.Module, Deployment: v.Key, Endpoint: targetEndpoint}
			modulesByName[v.Module] = v.Schema
		}
	}

	httpRoutes, err := s.dal.GetIngressRoutes(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get ingress routes: %w", err)
	}
	orderedModules := maps.Values(modulesByName)
	sort.SliceStable(orderedModules, func(i, j int) bool {
		return orderedModules[i].Name < orderedModules[j].Name
	})
	combined := &schema.Schema{Modules: orderedModules}
	s.schemaState.Store(schemaState{schema: combined, routes: newRoutes, httpRoutes: httpRoutes})
	return time.Second, nil
}

func (s *Service) reapCallEvents(ctx context.Context) (time.Duration, error) {
	logger := log.FromContext(ctx)

	if s.config.EventLogRetention == nil {
		logger.Tracef("Event log retention is disabled, will not prune.")
		return time.Hour, nil
	}

	removed, err := s.timeline.DeleteOldEvents(ctx, timeline.EventTypeCall, *s.config.EventLogRetention)
	if err != nil {
		return 0, fmt.Errorf("failed to prune call events: %w", err)
	}
	if removed > 0 {
		logger.Debugf("Pruned %d call events older than %s", removed, s.config.EventLogRetention)
	}

	// Prune every 5% of the retention period.
	return *s.config.EventLogRetention / 20, nil
}

func extractIngressRoutingEntries(req *ftlv1.CreateDeploymentRequest) []dal.IngressRoutingEntry {
	var ingressRoutes []dal.IngressRoutingEntry
	for _, decl := range req.Schema.Decls {
		if verb, ok := decl.Value.(*schemapb.Decl_Verb); ok {
			for _, metadata := range verb.Verb.Metadata {
				if ingress, ok := metadata.Value.(*schemapb.Metadata_Ingress); ok {
					ingressRoutes = append(ingressRoutes, dal.IngressRoutingEntry{
						Verb:   verb.Verb.Name,
						Method: ingress.Ingress.Method,
						Path:   ingressPathString(ingress.Ingress.Path),
					})
				}
			}
		}
	}
	return ingressRoutes
}

func ingressPathString(path []*schemapb.IngressPathComponent) string {
	pathString := make([]string, len(path))
	for i, p := range path {
		switch p.Value.(type) {
		case *schemapb.IngressPathComponent_IngressPathLiteral:
			pathString[i] = p.GetIngressPathLiteral().Text
		case *schemapb.IngressPathComponent_IngressPathParameter:
			pathString[i] = fmt.Sprintf("{%s}", p.GetIngressPathParameter().Name)
		}
	}
	return "/" + strings.Join(pathString, "/")
}

func makeBackoff(min, max time.Duration) backoff.Backoff {
	return backoff.Backoff{
		Min:    min,
		Max:    max,
		Jitter: true,
		Factor: 2,
	}
}

func validateCallBody(body []byte, verb *schema.Verb, sch *schema.Schema) error {
	var root any
	err := json.Unmarshal(body, &root)
	if err != nil {
		return fmt.Errorf("request body is not valid JSON: %w", err)
	}

	var opts []schema.EncodingOption
	if e, ok := slices.FindVariant[*schema.MetadataEncoding](verb.Metadata); ok && e.Lenient {
		opts = append(opts, schema.LenientMode())
	}
	err = schema.ValidateJSONValue(verb.Request, []string{verb.Request.String()}, root, sch, opts...)
	if err != nil {
		return fmt.Errorf("could not validate call request body: %w", err)
	}
	return nil
}

type Route struct {
	Module     string
	Deployment model.DeploymentKey
	Endpoint   string
}

type schemaState struct {
	schema     *schema.Schema
	routes     map[string]Route
	httpRoutes map[string][]dalmodel.IngressRoute
}

func (r Route) String() string {
	return fmt.Sprintf("%s -> %s", r.Deployment, r.Endpoint)
}
