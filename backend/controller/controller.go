package controller

import (
	"context"
	sha "crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
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

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/controller/state"
	ftldeployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	deploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/deploymentpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/runner/pubsub"
	"github.com/TBD54566975/ftl/backend/timeline"
	schemapb "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/deploymentcontext"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	ftlmaps "github.com/TBD54566975/ftl/internal/maps"
	"github.com/TBD54566975/ftl/internal/model"
	internalobservability "github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

// CommonConfig between the production controller and development server.
type CommonConfig struct {
	IdleRunners    int           `help:"Number of idle runners to keep around (not supported in production)." default:"3"`
	WaitFor        []string      `help:"Wait for these modules to be deployed before becoming ready." placeholder:"MODULE"`
	CronJobTimeout time.Duration `help:"Timeout for cron jobs." default:"5m"`
}

type Config struct {
	Bind                         *url.URL            `help:"Socket to bind to." default:"http://127.0.0.1:8892" env:"FTL_BIND"`
	Key                          model.ControllerKey `help:"Controller key (auto)." placeholder:"KEY"`
	DSN                          string              `help:"DAL DSN." default:"${dsn}" env:"FTL_DSN"`
	Advertise                    *url.URL            `help:"Endpoint the Controller should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_ADVERTISE"`
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
	adminClient ftlv1connect.AdminServiceClient,
	timelineClient *timeline.Client,
	conn *sql.DB,
	devel bool,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL controller")

	svc, err := New(ctx, conn, adminClient, timelineClient, storage, config, devel)
	if err != nil {
		return err
	}
	logger.Debugf("Listening on %s", config.Bind)
	logger.Debugf("Advertising as %s", config.Advertise)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return rpc.Serve(ctx, config.Bind,
			rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
			rpc.GRPC(deploymentconnect.NewDeploymentServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
			rpc.GRPC(ftlv1connect.NewSchemaServiceHandler, svc),
			rpc.PProf(),
		)
	})

	return g.Wait()
}

var _ ftlv1connect.ControllerServiceHandler = (*Service)(nil)
var _ ftlv1connect.SchemaServiceHandler = (*Service)(nil)

type clients struct {
	verb ftlv1connect.VerbServiceClient
}

type Service struct {
	leaser             leases.Leaser
	key                model.ControllerKey
	deploymentLogsSink *deploymentLogsSink
	adminClient        ftlv1connect.AdminServiceClient

	tasks          *scheduledtask.Scheduler
	pubSub         *pubsub.Service
	timelineClient *timeline.Client
	storage        *artefacts.OCIArtefactService

	// Map from runnerKey.String() to client.
	clients    *ttlcache.Cache[string, clients]
	clientLock sync.Mutex

	config Config

	routeTable      *routing.RouteTable
	controllerState state.ControllerState
}

func New(
	ctx context.Context,
	conn *sql.DB,
	adminClient ftlv1connect.AdminServiceClient,
	timelineClient *timeline.Client,
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

	routingTable := routing.New(ctx, schemaeventsource.New(ctx, rpc.ClientFromContext[ftlv1connect.SchemaServiceClient](ctx)))

	svc := &Service{
		tasks:           scheduler,
		timelineClient:  timelineClient,
		leaser:          ldb,
		key:             key,
		clients:         ttlcache.New(ttlcache.WithTTL[string, clients](time.Minute)),
		config:          config,
		routeTable:      routingTable,
		storage:         storage,
		controllerState: state.NewInMemoryState(),
		adminClient:     adminClient,
	}

	svc.deploymentLogsSink = newDeploymentLogsSink(ctx, timelineClient)

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

	singletonTask := func(job scheduledtask.Job, name string, maxNext, minDelay, maxDelay time.Duration, develBackoff ...backoff.Backoff) {
		maybeDevelJob, backoff := maybeDevelTask(job, name, maxNext, minDelay, maxDelay, develBackoff...)
		svc.tasks.Singleton(name, maybeDevelJob, backoff)
	}

	// Singleton tasks use leases to only run on a single controller.
	singletonTask(svc.reapStaleRunners, "reap-stale-runners", time.Second*2, time.Second, time.Second*10)
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
		minReplicas := int32(0)
		deployment, err := currentState.GetDeployment(p.Deployment)
		if err == nil {
			minReplicas = int32(deployment.MinReplicas)
		}
		return &ftlv1.ProcessListResponse_Process{
			Deployment:  p.Deployment.String(),
			Runner:      runner,
			MinReplicas: minReplicas,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.ProcessListResponse{Processes: out}), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	controller := state.Controller{Key: s.key, Endpoint: s.config.Bind.String()}
	currentState := s.controllerState.View()
	runners := currentState.Runners()
	status := currentState.GetActiveDeployments()
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
	deployments, err := slices.MapErr(maps.Values(status), func(d *state.Deployment) (*ftlv1.StatusResponse_Deployment, error) {
		return &ftlv1.StatusResponse_Deployment{
			Key:         d.Key.String(),
			Language:    d.Language,
			Name:        d.Module,
			MinReplicas: int32(d.MinReplicas),
			Replicas:    replicas[d.Key.String()],
			Schema:      d.Schema.ToProto(),
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

		s.timelineClient.Publish(ctx, timeline.Log{
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
	view := s.controllerState.View()
	schemas := view.GetActiveDeploymentSchemas()
	modules := []*schemapb.Module{
		schema.Builtins().ToProto(),
	}
	modules = append(modules, slices.Map(schemas, func(d *schema.Module) *schemapb.Module { return d.ToProto() })...)
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
	err = s.controllerState.Publish(ctx, &state.DeploymentSchemaUpdatedEvent{
		Key:    deployment,
		Schema: module,
	})
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
		err = s.setDeploymentReplicas(ctx, deploymentKey, int(*req.Msg.MinReplicas))
		if err != nil {
			logger.Errorf(err, "Could not set deployment replicas: %s", deploymentKey)
			return nil, fmt.Errorf("could not set deployment replicas: %w", err)
		}
	}
	return connect.NewResponse(&ftlv1.UpdateDeployResponse{}), nil
}

func (s *Service) setDeploymentReplicas(ctx context.Context, key model.DeploymentKey, minReplicas int) (err error) {

	view := s.controllerState.View()
	deployment, err := view.GetDeployment(key)
	if err != nil {
		return fmt.Errorf("could not get deployment: %w", err)
	}

	err = s.controllerState.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: key, Replicas: minReplicas})
	if err != nil {
		return fmt.Errorf("could not update deployment replicas: %w", err)
	}
	if minReplicas == 0 {
		err = s.controllerState.Publish(ctx, &state.DeploymentDeactivatedEvent{Key: key, ModuleRemoved: true})
		if err != nil {
			return fmt.Errorf("could not deactivate deployment: %w", err)
		}
	} else if deployment.MinReplicas == 0 {
		err = s.controllerState.Publish(ctx, &state.DeploymentActivatedEvent{Key: key, ActivatedAt: time.Now(), MinReplicas: minReplicas})
		if err != nil {
			return fmt.Errorf("could not activate deployment: %w", err)
		}
	}
	s.timelineClient.Publish(ctx, timeline.DeploymentUpdated{
		DeploymentKey:   key,
		MinReplicas:     minReplicas,
		PrevMinReplicas: deployment.MinReplicas,
	})

	return nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, c *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	newDeploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	logger := s.getDeploymentLogger(ctx, newDeploymentKey)
	logger.Debugf("Replace deployment for: %s", newDeploymentKey)

	view := s.controllerState.View()
	newDeployment, err := view.GetDeployment(newDeploymentKey)
	if err != nil {
		logger.Errorf(err, "Deployment not found: %s", newDeploymentKey)
		return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
	}
	minReplicas := int(c.Msg.MinReplicas)
	err = s.controllerState.Publish(ctx, &state.DeploymentActivatedEvent{Key: newDeploymentKey, ActivatedAt: time.Now(), MinReplicas: minReplicas})
	if err != nil {
		return nil, fmt.Errorf("replace deployment failed to activate: %w", err)
	}

	// If there's an existing deployment, set its desired replicas to 0
	var replacedDeploymentKey optional.Option[model.DeploymentKey]
	// TODO: remove all this, it needs to be event driven
	var oldDeployment *state.Deployment
	for _, dep := range view.GetActiveDeployments() {
		if dep.Module == newDeployment.Module {
			oldDeployment = dep
			break
		}
	}
	if oldDeployment != nil {
		if oldDeployment.Key.String() == newDeploymentKey.String() {
			return nil, fmt.Errorf("replace deployment failed: deployment already exists from %v to %v", oldDeployment.Key, newDeploymentKey)
		}
		err = s.controllerState.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: newDeploymentKey, Replicas: minReplicas})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to set new deployment replicas from %v to %v: %w", oldDeployment.Key, newDeploymentKey, err)
		}
		err = s.controllerState.Publish(ctx, &state.DeploymentDeactivatedEvent{Key: oldDeployment.Key})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to deactivate old deployment %v: %w", oldDeployment.Key, err)
		}
		replacedDeploymentKey = optional.Some(oldDeployment.Key)
	} else {
		// Set the desired replicas for the new deployment
		err = s.controllerState.Publish(ctx, &state.DeploymentReplicasUpdatedEvent{Key: newDeploymentKey, Replicas: minReplicas})
		if err != nil {
			return nil, fmt.Errorf("replace deployment failed to set replicas for %v: %w", newDeploymentKey, err)
		}
	}

	s.timelineClient.Publish(ctx, timeline.DeploymentCreated{
		DeploymentKey:      newDeploymentKey,
		Language:           newDeployment.Language,
		ModuleName:         newDeployment.Module,
		MinReplicas:        minReplicas,
		ReplacedDeployment: replacedDeploymentKey,
	})
	if err != nil {
		return nil, fmt.Errorf("replace deployment failed to create event: %w", err)
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
		Schema:    deployment.Schema.ToProto(),
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
		configsResp, err := s.adminClient.MapConfigsForModule(ctx, &connect.Request[ftlv1.MapConfigsForModuleRequest]{Msg: &ftlv1.MapConfigsForModuleRequest{Module: module}})
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get configs: %w", err))
		}
		configs := configsResp.Msg.Values
		routeTable := map[string]string{}
		for _, module := range callableModuleNames {
			deployment, ok := routeView.GetDeployment(module).Get()
			if !ok {
				continue
			}
			if route, ok := routeView.Get(deployment).Get(); ok {
				routeTable[deployment.String()] = route.String()
			}
		}
		if deployment.Schema.Runtime != nil && deployment.Schema.Runtime.Deployment != nil {
			routeTable[deployment.Key.String()] = deployment.Schema.Runtime.Deployment.Endpoint
		}

		secretsResp, err := s.adminClient.MapSecretsForModule(ctx, &connect.Request[ftlv1.MapSecretsForModuleRequest]{Msg: &ftlv1.MapSecretsForModuleRequest{Module: module}})
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("could not get secrets: %w", err))
		}
		secrets := secretsResp.Msg.Values

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

	module := verbRef.Module
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

	route, ok := routes.Get(deployment).Get()
	if !ok {
		err = fmt.Errorf("no routes for module %q", module)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("no routes for module"))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	if currentCaller != nil && currentCaller.Module != module && !verb.IsExported() {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: verb not exported"))
		err = connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("verb %q is not exported", verbRef))
	}

	err = validateCallBody(req.Msg.Body, verb, sch)
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("invalid request: invalid call body"))
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
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

	s.timelineClient.Publish(ctx, callEvent)
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

	dkey := model.NewDeploymentKey(module.Name)
	err = s.controllerState.Publish(ctx, &state.DeploymentCreatedEvent{
		Module:    module.Name,
		Key:       dkey,
		CreatedAt: time.Now(),
		Schema:    module,
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
	view := s.controllerState.View()
	existingModules := view.GetActiveDeployments()
	schemaMap := ftlmaps.FromSlice[string, *schema.Module, *state.Deployment](maps.Values(existingModules), func(el *state.Deployment) (string, *schema.Module) { return el.Module, el.Schema })
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

	updates := s.controllerState.Updates().Subscribe(nil)
	defer s.controllerState.Updates().Unsubscribe(updates)
	view := s.controllerState.View()

	// Seed the notification channel with the current deployments.
	seedDeployments := view.GetActiveDeployments()
	initialCount := len(seedDeployments)

	builtins := schema.Builtins().ToProto()
	builtinsResponse := &ftlv1.PullSchemaResponse{
		ModuleName: builtins.Name,
		Schema:     builtins,
		ChangeType: ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
		More:       initialCount > 0,
	}

	err := sendChange(builtinsResponse)
	if err != nil {
		return err
	}
	for _, initial := range seedDeployments {
		initialCount--
		module := initial.Schema.ToProto()
		err := sendChange(&ftlv1.PullSchemaResponse{
			ModuleName:    module.Name,
			DeploymentKey: proto.String(initial.Key.String()),
			Schema:        module,
			ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
			More:          initialCount > 0,
		})
		if err != nil {
			return err
		}
	}
	logger.Debugf("Seeded %d deployments", initialCount)

	for {
		select {
		case <-ctx.Done():
			return nil

		case notification := <-updates:
			switch event := notification.(type) {
			case *state.DeploymentCreatedEvent:
				err := sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
					ModuleName:    event.Module,
					DeploymentKey: proto.String(event.Key.String()),
					Schema:        event.Schema.ToProto(),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED,
				})
				if err != nil {
					return err
				}
			case *state.DeploymentDeactivatedEvent:
				view := s.controllerState.View()
				dep, err := view.GetDeployment(event.Key)
				if err != nil {
					logger.Errorf(err, "Deployment not found: %s", event.Key)
					continue
				}
				err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
					ModuleName:    dep.Module,
					DeploymentKey: proto.String(event.Key.String()),
					Schema:        dep.Schema.ToProto(),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED,
					ModuleRemoved: event.ModuleRemoved,
				})
				if err != nil {
					return err
				}
			case *state.DeploymentSchemaUpdatedEvent:
				view := s.controllerState.View()
				dep, err := view.GetDeployment(event.Key)
				if err != nil {
					logger.Errorf(err, "Deployment not found: %s", event.Key)
					continue
				}
				err = sendChange(&ftlv1.PullSchemaResponse{ //nolint:forcetypeassert
					ModuleName:    dep.Module,
					DeploymentKey: proto.String(event.Key.String()),
					Schema:        event.Schema.ToProto(),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED,
				})
				if err != nil {
					return err
				}
			default:
				continue
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
