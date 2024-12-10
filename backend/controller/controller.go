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
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/jackc/pgx/v5"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller/admin"
	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/console"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1/pbconsoleconnect"
	ftldeployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	deploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	ftlv1connect2 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	frontend "github.com/TBD54566975/ftl/frontend/console"
	"github.com/TBD54566975/ftl/internal/configuration"
	cf "github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/deploymentcontext"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	ftlmaps "github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	internalobservability "github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
	status "github.com/TBD54566975/ftl/internal/terminal"
)

// CommonConfig between the production controller and development server.
type CommonConfig struct {
	NoConsole      bool          `help:"Disable the console."`
	IdleRunners    int           `help:"Number of idle runners to keep around (not supported in production)." default:"3"`
	WaitFor        []string      `help:"Wait for these modules to be deployed before becoming ready." placeholder:"MODULE"`
	CronJobTimeout time.Duration `help:"Timeout for cron jobs." default:"5m"`
}

type Config struct {
	Bind                         *url.URL            `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Key                          model.ControllerKey `help:"Controller key (auto)." placeholder:"KEY"`
	DSN                          string              `help:"DAL DSN." default:"${dsn}" env:"FTL_CONTROLLER_DSN"`
	Advertise                    *url.URL            `help:"Endpoint the Controller should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_ADVERTISE"`
	ConsoleURL                   *url.URL            `help:"The public URL of the console (for CORS)." env:"FTL_CONTROLLER_CONSOLE_URL"`
	ContentTime                  time.Time           `help:"Time to use for console resource timestamps." default:"${timestamp=1970-01-01T00:00:00Z}"`
	RunnerTimeout                time.Duration       `help:"Runner heartbeat timeout." default:"10s"`
	ControllerTimeout            time.Duration       `help:"Controller heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration       `help:"Deployment reservation timeout." default:"120s"`
	ModuleUpdateFrequency        time.Duration       `help:"Frequency to send module updates." default:"30s"`
	ArtefactChunkSize            int                 `help:"Size of each chunk streamed to the client." default:"1048576"`
	MaxOpenDBConnections         int                 `help:"Maximum number of database connections." default:"20" env:"FTL_MAX_OPEN_DB_CONNECTIONS"`
	MaxIdleDBConnections         int                 `help:"Maximum number of idle database connections." default:"20" env:"FTL_MAX_IDLE_DB_CONNECTIONS"`
	CommonConfig
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c, kong.Vars{"dsn": dsn.PostgresDSN("ftl")}); err != nil {
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
	storage *artefacts.OCIArtefactService,
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

	svc, err := New(ctx, conn, cm, sm, storage, config, devel)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)

	admin := admin.NewAdminService(cm, sm, svc.dal)
	console := console.NewService(svc.dal, admin)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
			rpc.GRPC(deploymentconnect.NewDeploymentServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewAdminServiceHandler, admin),
			rpc.GRPC(pbconsoleconnect.NewConsoleServiceHandler, console),
			rpc.GRPC(timelinev1connect.NewTimelineServiceHandler, console),
			rpc.GRPC(ftlv1connect2.NewLegacyPubsubServiceHandler, svc.pubSub),
			rpc.HTTP("/", consoleHandler),
			rpc.PProf(),
		)
	})

	go svc.dal.PollDeployments(ctx)

	return g.Wait()
}

var _ ftlv1connect.ControllerServiceHandler = (*Service)(nil)
var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

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
	leaser             leases.Leaser
	dal                *dal.DAL
	key                model.ControllerKey
	deploymentLogsSink *deploymentLogsSink

	cm *cf.Manager[configuration.Configuration]
	sm *cf.Manager[configuration.Secrets]

	tasks                   *scheduledtask.Scheduler
	pubSub                  *pubsub.Service
	storage                 *artefacts.OCIArtefactService
	controllerListListeners []ControllerListListener

	// Map from runnerKey.String() to client.
	clients *ttlcache.Cache[string, clients]

	schemaSyncLock sync.Mutex

	config Config

	increaseReplicaFailures map[string]int

	clientLock      sync.Mutex
	routeTable      *routing.RouteTable
	controllerState state.ControllerState
}

func New(
	ctx context.Context,
	conn *sql.DB,
	cm *cf.Manager[configuration.Configuration],
	sm *cf.Manager[configuration.Secrets],
	storage *artefacts.OCIArtefactService,
	config Config,
	devel bool,
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

	ldb := leases.NewClientLeaser(ctx)
	scheduler := scheduledtask.New(ctx, key, ldb)

	routingTable := routing.New(ctx, schemaeventsource.New(ctx, rpc.Dial[ftlv1connect.SchemaServiceClient](ftlv1connect.NewSchemaServiceClient, config.Bind.String(), log.Error)))

	svc := &Service{
		cm:                      cm,
		sm:                      sm,
		tasks:                   scheduler,
		leaser:                  ldb,
		conn:                    conn,
		key:                     key,
		clients:                 ttlcache.New(ttlcache.WithTTL[string, clients](time.Minute)),
		config:                  config,
		increaseReplicaFailures: map[string]int{},
		routeTable:              routingTable,
		storage:                 storage,
		controllerState:         state.NewInMemoryState(),
	}

	pubSub := pubsub.New(ctx, conn, routingTable, svc.controllerState)
	svc.pubSub = pubSub
	svc.dal = dal.New(ctx, conn, svc.storage, svc.controllerState)

	svc.deploymentLogsSink = newDeploymentLogsSink(ctx)

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
	parallelTask(svc.pubSub.ExecuteAsyncCalls, "execute-async-calls", time.Second, time.Second*5, time.Second*10)

	// Singleton tasks use leases to only run on a single controller.
	singletonTask(svc.reapStaleRunners, "reap-stale-runners", time.Second*2, time.Second, time.Second*10)
	singletonTask(svc.pubSub.ReapAsyncCalls, "reap-async-calls", time.Second*5, time.Second, time.Second*5)
	return svc, nil
}

func (s *Service) ProcessList(ctx context.Context, req *connect.Request[ftlv1.ProcessListRequest]) (*connect.Response[ftlv1.ProcessListResponse], error) {
	currentState := s.controllerState.View()
	runners := currentState.Runners()

	out, err := slices.MapErr(runners, func(p state.Runner) (*ftlv1.ProcessListResponse_Process, error) {
		runner := &ftlv1.ProcessListResponse_ProcessRunner{
			Key:      p.Key.String(),
			Endpoint: p.Endpoint,
		}
		return &ftlv1.ProcessListResponse_Process{
			Deployment: p.Deployment.String(),
			Runner:     runner,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.ProcessListResponse{Processes: out}), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	controller := dalmodel.Controller{Key: s.key, Endpoint: s.config.Bind.String()}
	currentState := s.controllerState.View()
	runners := currentState.Runners()
	status, err := s.dal.GetActiveDeployments()
	if err != nil {
		return nil, fmt.Errorf("could not get status: %w", err)
	}
	allModules := s.routeTable.Current()
	routes := slices.Map(allModules.Schema().Modules, func(module *schema.Module) (out *ftlv1.StatusResponse_Route) {
		key := ""
		endpoint := ""
		if module.Runtime != nil && module.Runtime.Deployment != nil {
			key = module.Runtime.Deployment.DeploymentKey
			endpoint = module.Runtime.Deployment.Endpoint
		}
		return &ftlv1.StatusResponse_Route{
			Module:     module.Name,
			Deployment: key,
			Endpoint:   endpoint,
		}
	})
	replicas := map[string]int32{}
	protoRunners, err := slices.MapErr(runners, func(r state.Runner) (*ftlv1.StatusResponse_Runner, error) {
		asString := r.Deployment.String()
		deployment := &asString
		replicas[asString]++
		return &ftlv1.StatusResponse_Runner{
			Key:        r.Key.String(),
			Endpoint:   r.Endpoint,
			Deployment: deployment,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	deployments, err := slices.MapErr(status, func(d dalmodel.Deployment) (*ftlv1.StatusResponse_Deployment, error) {
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
		Controllers: []*ftlv1.StatusResponse_Controller{{
			Key:      controller.Key.String(),
			Endpoint: controller.Endpoint,
			Version:  ftl.Version,
		}},
		Runners:     protoRunners,
		Deployments: deployments,
		Routes:      routes,
	}
	return connect.NewResponse(resp), nil
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

		timeline.ClientFromContext(ctx).Publish(ctx, timeline.Log{
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

func (s *Service) UpdateDeploymentRuntime(ctx context.Context, req *connect.Request[ftlv1.UpdateDeploymentRuntimeRequest]) (*connect.Response[ftlv1.UpdateDeploymentRuntimeResponse], error) {
	deployment, err := model.ParseDeploymentKey(req.Msg.Deployment)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	cs := s.controllerState.View()
	dep, err := cs.GetDeployment(deployment)
	if err != nil {
		return nil, fmt.Errorf("could not get schema: %w", err)
	}
	module := dep.Schema
	if module.Runtime == nil {
		module.Runtime = &schema.ModuleRuntime{}
	}
	event := schema.ModuleRuntimeEventFromProto(req.Msg.Event)
	module.Runtime.ApplyEvent(event)
	err = s.dal.UpdateModuleSchema(ctx, deployment, module)
	if err != nil {
		return nil, fmt.Errorf("could not update schema for module %s: %w", module.Name, err)
	}

	return connect.NewResponse(&ftlv1.UpdateDeploymentRuntimeResponse{}), nil
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (response *connect.Response[ftlv1.UpdateDeployResponse], err error) {
	deploymentKey, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}

	logger := s.getDeploymentLogger(ctx, deploymentKey)
	logger.Debugf("Update deployment for: %s", deploymentKey)
	if req.Msg.MinReplicas != nil {
		err = s.dal.SetDeploymentReplicas(ctx, deploymentKey, int(*req.Msg.MinReplicas))
		if err != nil {
			if errors.Is(err, libdal.ErrNotFound) {
				logger.Errorf(err, "Deployment not found: %s", deploymentKey)
				return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
			}
			logger.Errorf(err, "Could not set deployment replicas: %s", deploymentKey)
			return nil, fmt.Errorf("could not set deployment replicas: %w", err)
		}
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
			cs := s.controllerState.View()
			dep, err := cs.GetDeployment(newDeploymentKey)
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
		// The created event does not matter if it is a new runner or not.
		err = s.controllerState.Publish(ctx, &state.RunnerRegisteredEvent{
			Key:        runnerKey,
			Endpoint:   msg.Endpoint,
			Deployment: deploymentKey,
			Time:       time.Now(),
		})
		if err != nil {
			return nil, err
		}
		if !deferredDeregistration {
			// Deregister the runner if the Runner disconnects.
			defer func() {
				err := s.controllerState.Publish(ctx, &state.RunnerDeletedEvent{Key: runnerKey})
				if err != nil {
					logger.Errorf(err, "Could not deregister runner %s", runnerStr)
				}
			}()
			deferredDeregistration = true
		}
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
		Artefacts: slices.Map(maps.Values(deployment.Artefacts), ftlv1.ArtefactToProto),
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
		reader, err := s.storage.Download(ctx, artefact.Digest)
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

	routeView := s.routeTable.Current()
	// It's not actually ready until it is in the routes table
	var missing []string
	for _, module := range s.config.WaitFor {
		if _, ok := routeView.GetForModule(module).Get(); !ok {
			missing = append(missing, module)
		}
	}
	if len(missing) == 0 {
		return connect.NewResponse(&ftlv1.PingResponse{}), nil
	}

	msg := fmt.Sprintf("waiting for deployments: %s", strings.Join(missing, ", "))
	return connect.NewResponse(&ftlv1.PingResponse{NotReady: &msg}), nil
}

// GetDeploymentContext retrieves config, secrets and DSNs for a module.
func (s *Service) GetDeploymentContext(ctx context.Context, req *connect.Request[ftldeployment.GetDeploymentContextRequest], resp *connect.ServerStream[ftldeployment.GetDeploymentContextResponse]) error {

	logger := log.FromContext(ctx)
	updates := s.routeTable.Subscribe()
	defer s.routeTable.Unsubscribe(updates)
	depName := req.Msg.Deployment
	if !strings.HasPrefix(depName, "dpl-") {
		// For hot reload endponts we might not have a deployment key
		deps, err := s.dal.GetActiveDeployments()
		if err != nil {
			return fmt.Errorf("could not get active deployments: %w", err)
		}
		for _, dep := range deps {
			if dep.Module == depName {
				depName = dep.Key.String()
				break
			}
		}
	}
	key, err := model.ParseDeploymentKey(depName)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	cs := s.controllerState.View()
	deployment, err := cs.GetDeployment(key)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}
	module := deployment.Module

	// Initialize checksum to -1; a zero checksum does occur when the context contains no settings
	lastChecksum := int64(-1)

	callableModules := map[string]bool{}
	for _, decl := range deployment.Schema.Decls {
		switch entry := decl.(type) {
		case *schema.Verb:
			for _, md := range entry.Metadata {
				if calls, ok := md.(*schema.MetadataCalls); ok {
					for _, call := range calls.Calls {
						callableModules[call.Module] = true
					}
				}
			}
		default:

		}
	}
	callableModuleNames := maps.Keys(callableModules)
	callableModuleNames = slices.Sort(callableModuleNames)
	logger.Debugf("Modules %s can call %v", module, callableModuleNames)
	for {
		h := sha.New()

		routeView := s.routeTable.Current()
		configs, err := s.cm.MapForModule(ctx, module)
		routeTable := map[string]string{}
		for _, module := range callableModuleNames {
			if route, ok := routeView.GetForModule(module).Get(); ok {
				routeTable[module] = route.String()
			}
		}
		if deployment.Schema.Runtime != nil && deployment.Schema.Runtime.Deployment != nil {
			routeTable[module] = deployment.Schema.Runtime.Deployment.Endpoint
		}

		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get configs: %w", err))
		}
		secrets, err := s.sm.MapForModule(ctx, module)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get secrets: %w", err))
		}

		if err := hashConfigurationMap(h, configs); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on configs: %w", err))
		}
		if err := hashConfigurationMap(h, secrets); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on secrets: %w", err))
		}
		if err := hashRoutesTable(h, routeTable); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not detect change on routes: %w", err))
		}

		checksum := int64(binary.BigEndian.Uint64((h.Sum(nil))[0:8]))

		if checksum != lastChecksum {
			logger.Debugf("Sending module context for: %s routes: %v", module, routeTable)
			response := deploymentcontext.NewBuilder(module).AddConfigs(configs).AddSecrets(secrets).AddRoutes(routeTable).Build().ToProto()

			if err := resp.Send(response); err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("could not send response: %w", err))
			}

			lastChecksum = checksum
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(s.config.ModuleUpdateFrequency):
		case <-updates:

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

// hashRoutesTable computes an order invariant checksum on the routes
func hashRoutesTable(h hash.Hash, m map[string]string) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return fmt.Errorf("error hashing routes: %w", err)
		}
	}
	return nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	return s.callWithRequest(ctx, headers.CopyRequestForForwarding(req), optional.None[model.RequestKey](), optional.None[model.RequestKey](), "")
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

	routes := s.routeTable.Current()
	sch := routes.Schema()

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
	route, ok := routes.GetForModule(module).Get()
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
	}

	deployment, ok := routes.GetDeployment(module).Get()
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment"))
		return nil, fmt.Errorf("deployment not found for module %q", module)
	}
	callEvent := &timeline.Call{
		DeploymentKey:    deployment,
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
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		timeline.ClientFromContext(ctx).Publish(ctx, callEvent)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
	}

	err = validateCallBody(req.Msg.Body, verb, sch)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid call body"))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		timeline.ClientFromContext(ctx).Publish(ctx, callEvent)
		return nil, err
	}

	client := s.clientsForEndpoint(route.String())

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
		callEvent.Response = result.Ok(resp.Msg)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	} else {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		logger.Errorf(err, "Call failed to verb %s for module %s", verbRef.String(), module)
	}

	timeline.ClientFromContext(ctx).Publish(ctx, callEvent)
	return resp, err
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := slices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, err
	}
	_, need, err := s.storage.GetDigestsKeys(ctx, byteDigests)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: slices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx)
	digest, err := s.storage.Upload(ctx, artefacts.Artefact{Content: req.Msg.Content})
	if err != nil {
		return nil, err
	}
	logger.Debugf("Created new artefact %s", digest)
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: digest[:]}), nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)

	artefacts := make([]*state.DeploymentArtefact, len(req.Msg.Artefacts))
	for i, artefact := range req.Msg.Artefacts {
		digest, err := sha256.ParseSHA256(artefact.Digest)
		if err != nil {
			logger.Errorf(err, "Invalid digest %s", artefact.Digest)
			return nil, fmt.Errorf("invalid digest: %w", err)
		}
		err = s.controllerState.Publish(ctx, &state.DeploymentArtefactCreatedEvent{})
		if err != nil {
			return nil, fmt.Errorf("could not create deployment artefact: %w", err)
		}
		artefacts[i] = &state.DeploymentArtefact{
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

	dkey, err := s.dal.CreateDeployment(ctx, ms.Runtime.Base.Language, module)
	if err != nil {
		logger.Errorf(err, "Could not create deployment")
		return nil, fmt.Errorf("could not create deployment: %w", err)
	}
	err = s.controllerState.Publish(ctx, &state.DeploymentCreatedEvent{
		Module:    module.Name,
		Key:       dkey,
		CreatedAt: time.Now(),
		Schema:    ms,
		Artefacts: artefacts,
		Language:  ms.Runtime.Base.Language,
	})
	if err != nil {
		logger.Errorf(err, "Could not create deployment event")
		return nil, fmt.Errorf("could not create deployment event: %w", err)
	}

	deploymentLogger := s.getDeploymentLogger(ctx, dkey)
	deploymentLogger.Debugf("Created deployment %s", dkey)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: dkey.String()}), nil
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

func (s *Service) getDeployment(ctx context.Context, key string) (*state.Deployment, error) {
	dkey, err := model.ParseDeploymentKey(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid deployment key: %w", err))
	}
	cs := s.controllerState.View()
	deployment, err := cs.GetDeployment(dkey)
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
	cs := s.controllerState.View()

	for _, runner := range cs.Runners() {
		if runner.LastSeen.Add(s.config.RunnerTimeout).Before(time.Now()) {
			runnerKey := runner.Key
			logger.Debugf("Reaping stale runner %s with last seen %v", runnerKey, runner.LastSeen.String())
			if err := s.controllerState.Publish(ctx, &state.RunnerDeletedEvent{Key: runnerKey}); err != nil {
				return 0, fmt.Errorf("failed to publish runner deleted event: %w", err)
			}
		}
	}
	return s.config.RunnerTimeout, nil
}

func (s *Service) watchModuleChanges(ctx context.Context, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx)
	type moduleStateEntry struct {
		hash        []byte
		minReplicas int
	}
	deploymentState := map[string]moduleStateEntry{}
	moduleByDeploymentKey := map[string]string{}
	aliveDeploymentsForModule := map[string]map[string]bool{}
	schemaByDeploymentKey := map[string]*schemapb.Module{}

	// Seed the notification channel with the current deployments.
	seedDeployments, err := s.dal.GetActiveDeployments()
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
	builtinsResponse := &ftlv1.PullSchemaResponse{
		ModuleName: builtins.Name,
		Schema:     builtins,
		ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
		More:       initialCount > 0,
	}

	err = sendChange(builtinsResponse)
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
				schema := schemaByDeploymentKey[deletion.String()]
				moduleRemoved := true
				if aliveDeploymentsForModule[name] != nil {
					delete(aliveDeploymentsForModule[name], deletion.String())
					moduleRemoved = len(aliveDeploymentsForModule[name]) == 0
					if moduleRemoved {
						delete(aliveDeploymentsForModule, name)
					}
				}
				response = &ftlv1.PullSchemaResponse{
					ModuleName:    name,
					DeploymentKey: proto.String(deletion.String()),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED,
					ModuleRemoved: moduleRemoved,
					Schema:        schema,
				}
				delete(deploymentState, deletion.String())
				delete(moduleByDeploymentKey, deletion.String())
				delete(schemaByDeploymentKey, deletion.String())
			} else if message, ok := notification.Message.Get(); ok {
				if message.Schema.Runtime == nil {
					message.Schema.Runtime = &schema.ModuleRuntime{}
				}
				message.Schema.Runtime.Scaling = &schema.ModuleRuntimeScaling{
					MinReplicas: int32(message.MinReplicas),
				}
				message.Schema.Runtime.Base.CreateTime = message.CreatedAt

				moduleSchema := message.Schema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
				hasher := sha.New()
				data, err := schema.ModuleToBytes(message.Schema)
				if err != nil {
					logger.Errorf(err, "Could not serialize module schema")
					return fmt.Errorf("could not serialize module schema: %w", err)
				}
				if _, err := hasher.Write(data); err != nil {
					return err
				}

				newState := moduleStateEntry{
					hash:        hasher.Sum(nil),
					minReplicas: message.MinReplicas,
				}
				if current, ok := deploymentState[message.Key.String()]; ok {
					if !bytes.Equal(current.hash, newState.hash) || current.minReplicas != newState.minReplicas {
						alive := aliveDeploymentsForModule[moduleSchema.Name]
						if alive == nil {
							alive = map[string]bool{}
							aliveDeploymentsForModule[moduleSchema.Name] = alive
						}
						if newState.minReplicas > 0 {
							alive[message.Key.String()] = true
						} else {
							delete(alive, message.Key.String())
						}
						changeType := ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED
						// A deployment is considered removed if its minReplicas is set to 0.
						moduleRemoved := false
						if current.minReplicas > 0 && message.MinReplicas == 0 {
							changeType = ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED
							moduleRemoved = len(alive) == 0
							logger.Infof("Deployment %s was deleted via update notfication with module removed %v", deletion, moduleRemoved)
						}
						response = &ftlv1.PullSchemaResponse{
							ModuleName:    moduleSchema.Name,
							DeploymentKey: proto.String(message.Key.String()),
							Schema:        moduleSchema,
							ChangeType:    changeType,
							ModuleRemoved: moduleRemoved,
						}
					}
				} else {
					alive := aliveDeploymentsForModule[moduleSchema.Name]
					if alive == nil {
						alive = map[string]bool{}
						aliveDeploymentsForModule[moduleSchema.Name] = alive
					}
					alive[message.Key.String()] = true
					response = &ftlv1.PullSchemaResponse{
						ModuleName:    moduleSchema.Name,
						DeploymentKey: proto.String(message.Key.String()),
						Schema:        moduleSchema,
						ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
						More:          initialCount > 1,
					}
					if initialCount > 0 {
						initialCount--
					}
				}
				deploymentState[message.Key.String()] = newState
				moduleByDeploymentKey[message.Key.String()] = message.Schema.Name
				schemaByDeploymentKey[message.Key.String()] = moduleSchema
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
