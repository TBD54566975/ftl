package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jellydator/ttlcache/v3"
	"github.com/jpillora/backoff"
	"github.com/oklog/ulid/v2"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/cors"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/backend/common/rpc/headers"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	"github.com/TBD54566975/ftl/backend/schema"
	frontend "github.com/TBD54566975/ftl/frontend"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Config struct {
	Bind                         *url.URL            `help:"Socket to bind to." default:"http://localhost:8892" env:"FTL_CONTROLLER_BIND"`
	Advertise                    *url.URL            `help:"Endpoint the Controller should advertise (must be unique across the cluster, defaults to --bind if omitted)." env:"FTL_CONTROLLER_ADVERTISE"`
	ConsoleURL                   *url.URL            `help:"The public URL of the console (for CORS)." env:"FTL_CONTROLLER_CONSOLE_URL"`
	AllowOrigins                 []*url.URL          `help:"Allow CORS requests to ingress endpoints from these origins." env:"FTL_CONTROLLER_ALLOW_ORIGIN"`
	ContentTime                  time.Time           `help:"Time to use for console resource timestamps." default:"${timestamp=1970-01-01T00:00:00Z}"`
	Key                          model.ControllerKey `help:"Controller key (auto)." placeholder:"C<ULID>" default:"C00000000000000000000000000"`
	DSN                          string              `help:"DAL DSN." default:"postgres://localhost:54320/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_CONTROLLER_DSN"`
	RunnerTimeout                time.Duration       `help:"Runner heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration       `help:"Deployment reservation timeout." default:"120s"`
	ArtefactChunkSize            int                 `help:"Size of each chunk streamed to the client." default:"1048576"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
	if c.Advertise == nil {
		c.Advertise = c.Bind
	}
}

// Start the Controller. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config, runnerScaling scaling.RunnerScaling) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL controller")

	c, err := frontend.Server(ctx, config.ContentTime, config.ConsoleURL)
	if err != nil {
		return errors.WithStack(err)
	}

	// Bring up the DB connection and DAL.
	conn, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return errors.WithStack(err)
	}
	dal, err := dal.New(ctx, conn)
	if err != nil {
		return errors.WithStack(err)
	}

	svc, err := New(ctx, dal, config, runnerScaling)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Listening on %s", config.Bind)

	console := NewConsoleService(dal)

	ingressHandler := http.StripPrefix("/ingress", svc)
	if len(config.AllowOrigins) > 0 {
		ingressHandler = cors.Middleware(slices.Map(config.AllowOrigins, func(u *url.URL) string { return u.String() }), ingressHandler)
	}

	return rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
		rpc.GRPC(pbconsoleconnect.NewConsoleServiceHandler, console),
		rpc.HTTP("/ingress/", ingressHandler),
		rpc.HTTP("/", c),
	)
}

var _ ftlv1connect.ControllerServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type clients struct {
	verb   ftlv1connect.VerbServiceClient
	runner ftlv1connect.RunnerServiceClient
}

type Service struct {
	dal                *dal.DAL
	key                model.ControllerKey
	deploymentLogsSink *deploymentLogsSink

	// Map from endpoint to client.
	clients *ttlcache.Cache[string, clients]

	routesMu      sync.RWMutex
	routes        map[string][]dal.Route
	config        Config
	runnerScaling scaling.RunnerScaling
}

func New(ctx context.Context, db *dal.DAL, config Config, runnerScaling scaling.RunnerScaling) (*Service, error) {
	key := config.Key
	if config.Key.ULID() == (ulid.ULID{}) {
		key = model.NewControllerKey()
	}
	config.SetDefaults()
	svc := &Service{
		dal:                db,
		key:                key,
		deploymentLogsSink: newDeploymentLogsSink(ctx, db),
		clients:            ttlcache.New[string, clients](ttlcache.WithTTL[string, clients](time.Minute)),
		routes:             map[string][]dal.Route{},
		config:             config,
		runnerScaling:      runnerScaling,
	}

	go runWithRetries(ctx, time.Second*1, time.Second*2, svc.syncRoutes)
	go runWithRetries(ctx, time.Second*3, time.Second*5, svc.heartbeatController)
	go runWithRetries(ctx, time.Second*10, time.Second*20, svc.reapStaleControllers)
	go runWithRetries(ctx, config.RunnerTimeout, time.Second*10, svc.reapStaleRunners)
	go runWithRetries(ctx, config.DeploymentReservationTimeout, time.Second*20, svc.releaseExpiredReservations)
	go runWithRetries(ctx, time.Second*1, time.Second*5, svc.reconcileDeployments)
	go runWithRetries(ctx, time.Second*1, time.Second*5, svc.reconcileRunners)
	return svc, nil
}

// ServeHTTP handles ingress routes.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	logger.Infof("%s %s", r.Method, r.URL.Path)
	routes, err := s.dal.GetIngressRoutes(r.Context(), r.Method, r.URL.Path)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	route := routes[rand.Intn(len(routes))] //nolint:gosec
	var body []byte
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		// TODO: Transcode query parameters into JSON.
		payload := map[string]string{}
		for key, value := range r.URL.Query() {
			payload[key] = value[len(value)-1]
		}
		body, err = json.Marshal(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	requestName, err := s.dal.CreateIngressRequest(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path), r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	creq := connect.NewRequest(&ftlv1.CallRequest{
		Verb: &schemapb.VerbRef{Module: route.Module, Name: route.Verb},
		Body: body,
	})
	headers.SetRequestName(creq.Header(), requestName)
	resp, err := s.Call(r.Context(), creq)
	if err != nil {
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			http.Error(w, err.Error(), connectCodeToHTTP(connectErr.Code()))
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	switch msg := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Body:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(msg.Body)

	case *ftlv1.CallResponse_Error_:
		http.Error(w, msg.Error.Message, http.StatusInternalServerError)
	}
}

func (s *Service) ProcessList(ctx context.Context, req *connect.Request[ftlv1.ProcessListRequest]) (*connect.Response[ftlv1.ProcessListResponse], error) {
	processes, err := s.dal.GetProcessList(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out, err := slices.MapErr(processes, func(p dal.Process) (*ftlv1.ProcessListResponse_Process, error) {
		var runner *ftlv1.ProcessListResponse_ProcessRunner
		if dbr, ok := p.Runner.Get(); ok {
			labels, err := structpb.NewStruct(dbr.Labels)
			if err != nil {
				return nil, errors.Wrapf(err, "could not marshal labels for runner %s", dbr.Key)
			}
			runner = &ftlv1.ProcessListResponse_ProcessRunner{
				Key:      dbr.Key.String(),
				Endpoint: dbr.Endpoint,
				Labels:   labels,
			}
		}
		labels, err := structpb.NewStruct(p.Labels)
		if err != nil {
			return nil, errors.Wrapf(err, "could not marshal labels for deployment %s", p.Deployment)
		}
		return &ftlv1.ProcessListResponse_Process{
			Deployment:  p.Deployment.String(),
			MinReplicas: int32(p.MinReplicas),
			Labels:      labels,
			Runner:      runner,
		}, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&ftlv1.ProcessListResponse{Processes: out}), nil
}

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	status, err := s.dal.GetStatus(ctx, req.Msg.AllControllers, req.Msg.AllRunners, req.Msg.AllDeployments, req.Msg.AllIngressRoutes)
	if err != nil {
		return nil, errors.Wrap(err, "could not get status")
	}
	s.routesMu.RLock()
	routes := slices.FlatMap(maps.Values(s.routes), func(routes []dal.Route) (out []*ftlv1.StatusResponse_Route) {
		out = make([]*ftlv1.StatusResponse_Route, len(routes))
		for i, route := range routes {
			out[i] = &ftlv1.StatusResponse_Route{
				Module:     route.Module,
				Runner:     route.Runner.String(),
				Deployment: route.Deployment.String(),
				Endpoint:   route.Endpoint,
			}
		}
		return out
	})
	s.routesMu.RUnlock()
	protoRunners, err := slices.MapErr(status.Runners, func(r dal.Runner) (*ftlv1.StatusResponse_Runner, error) {
		var deployment *string
		if d, ok := r.Deployment.Get(); ok {
			asString := d.String()
			deployment = &asString
		}
		labels, err := structpb.NewStruct(r.Labels)
		if err != nil {
			return nil, errors.Wrapf(err, "could not marshal attributes for runner %s", r.Key)
		}
		return &ftlv1.StatusResponse_Runner{
			Key:        r.Key.String(),
			Endpoint:   r.Endpoint,
			State:      r.State.ToProto(),
			Deployment: deployment,
			Labels:     labels,
		}, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	deployments, err := slices.MapErr(status.Deployments, func(d dal.Deployment) (*ftlv1.StatusResponse_Deployment, error) {
		labels, err := structpb.NewStruct(d.Labels)
		if err != nil {
			return nil, errors.Wrapf(err, "could not marshal attributes for deployment %s", d.Name)
		}
		return &ftlv1.StatusResponse_Deployment{
			Key:         d.Name.String(),
			Language:    d.Language,
			Name:        d.Module,
			MinReplicas: int32(d.MinReplicas),
			Schema:      d.Schema.ToProto().(*schemapb.Module), //nolint:forcetypeassert
			Labels:      labels,
		}, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp := &ftlv1.StatusResponse{
		Controllers: slices.Map(status.Controllers, func(c dal.Controller) *ftlv1.StatusResponse_Controller {
			return &ftlv1.StatusResponse_Controller{
				Key:      c.Key.String(),
				Endpoint: c.Endpoint,
				State:    c.State.ToProto(),
			}
		}),
		Runners:     protoRunners,
		Deployments: deployments,
		IngressRoutes: slices.Map(status.IngressRoutes, func(r dal.IngressRouteEntry) *ftlv1.StatusResponse_IngressRoute {
			return &ftlv1.StatusResponse_IngressRoute{
				DeploymentName: r.Deployment.String(),
				Verb:           &schemapb.VerbRef{Module: r.Module, Name: r.Verb},
				Method:         r.Method,
				Path:           r.Path,
			}
		}),
		Routes: routes,
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) StreamDeploymentLogs(ctx context.Context, stream *connect.ClientStream[ftlv1.StreamDeploymentLogsRequest]) (*connect.Response[ftlv1.StreamDeploymentLogsResponse], error) {
	for stream.Receive() {
		msg := stream.Msg()
		deploymentName, err := model.ParseDeploymentName(msg.DeploymentName)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
		}
		var requestName types.Option[model.RequestName]
		if msg.RequestName != nil {
			_, rkey, err := model.ParseRequestName(*msg.RequestName)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid request key"))
			}
			requestName = types.Some(rkey)
		}

		err = s.dal.InsertLogEvent(ctx, &dal.LogEvent{
			RequestName:    requestName,
			DeploymentName: deploymentName,
			Time:           msg.TimeStamp.AsTime(),
			Level:          msg.LogLevel,
			Attributes:     msg.Attributes,
			Message:        msg.Message,
			Error:          types.Ptr(msg.Error),
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if stream.Err() != nil {
		return nil, errors.WithStack(stream.Err())
	}
	return connect.NewResponse(&ftlv1.StreamDeploymentLogsResponse{}), nil
}

func (s *Service) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	deployments, err := s.dal.GetActiveDeployments(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sch := &schemapb.Schema{
		Modules: slices.Map(deployments, func(d dal.Deployment) *schemapb.Module {
			return d.Schema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
		}),
	}
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: sch}), nil
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return s.watchModuleChanges(ctx, func(response *ftlv1.PullSchemaResponse) error {
		return errors.WithStack(stream.Send(response))
	})
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (response *connect.Response[ftlv1.UpdateDeployResponse], err error) {
	deploymentName, err := model.ParseDeploymentName(req.Msg.DeploymentName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	logger := s.getDeploymentLogger(ctx, deploymentName)
	logger.Infof("Update deployment for: %s", deploymentName)

	err = s.dal.SetDeploymentReplicas(ctx, deploymentName, int(req.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			logger.Errorf(err, "Deployment not found: %s", deploymentName)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		}
		logger.Errorf(err, "Could not set deployment replicas: %s", deploymentName)
		return nil, errors.Wrap(err, "could not set deployment replicas")
	}

	return connect.NewResponse(&ftlv1.UpdateDeployResponse{}), nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, c *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	newDeploymentName, err := model.ParseDeploymentName(c.Msg.DeploymentName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
	}

	logger := s.getDeploymentLogger(ctx, newDeploymentName)
	logger.Infof("Replace deployment for: %s", newDeploymentName)

	err = s.dal.ReplaceDeployment(ctx, newDeploymentName, int(c.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			logger.Errorf(err, "Deployment not found: %s", newDeploymentName)
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		} else if errors.Is(err, dal.ErrConflict) {
			logger.Infof("Deployment already exists: %s", newDeploymentName)
		} else {
			logger.Errorf(err, "Could not replace deployment: %s", newDeploymentName)
			return nil, errors.Wrap(err, "could not replace deployment")
		}
	}
	return connect.NewResponse(&ftlv1.ReplaceDeployResponse{}), nil
}

func (s *Service) RegisterRunner(ctx context.Context, stream *connect.ClientStream[ftlv1.RegisterRunnerRequest]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {
	initialised := false

	logger := log.FromContext(ctx)
	for stream.Receive() {
		msg := stream.Msg()
		endpoint, err := url.Parse(msg.Endpoint)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid endpoint"))
		}
		if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("invalid endpoint scheme %q", endpoint.Scheme))
		}
		runnerKey, err := model.ParseRunnerKey(msg.Key)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid key"))
		}

		runnerStr := fmt.Sprintf("%s (%s)", endpoint, runnerKey)
		logger.Tracef("Heartbeat received from runner %s", runnerStr)

		if !initialised {
			// Deregister the runner if the Runner disconnects.
			defer func() {
				err := s.dal.DeregisterRunner(context.Background(), runnerKey)
				if err != nil {
					logger.Errorf(err, "Could not deregister runner %s", runnerStr)
				}
			}()
			err = s.pingRunner(ctx, endpoint)
			if err != nil {
				return nil, errors.Wrap(err, "runner callback failed")
			}
			initialised = true
		}

		maybeDeployment, err := msg.DeploymentAsOptional()
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
		}
		err = s.dal.UpsertRunner(ctx, dal.Runner{
			Key:        runnerKey,
			Endpoint:   msg.Endpoint,
			State:      dal.RunnerStateFromProto(msg.State),
			Deployment: maybeDeployment,
			Labels:     msg.Labels.AsMap(),
		})
		if errors.Is(err, dal.ErrConflict) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if stream.Err() != nil {
		return nil, errors.WithStack(stream.Err())
	}
	return connect.NewResponse(&ftlv1.RegisterRunnerResponse{}), nil
}

// Check if we can contact the runner.
func (s *Service) pingRunner(ctx context.Context, endpoint *url.URL) error {
	client := rpc.Dial(ftlv1connect.NewRunnerServiceClient, endpoint.String(), log.Error)
	retry := backoff.Backoff{}
	heartbeatCtx, cancel := context.WithTimeout(ctx, s.config.RunnerTimeout)
	defer cancel()
	err := rpc.Wait(heartbeatCtx, retry, client)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to connect to runner"))
	}
	return nil
}

func (s *Service) GetDeployment(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentName)
	if err != nil {
		return nil, err
	}

	logger := s.getDeploymentLogger(ctx, deployment.Name)
	logger.Infof("Get deployment for: %s", deployment.Name)

	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema:    deployment.Schema.ToProto().(*schemapb.Module), //nolint:forcetypeassert
		Artefacts: slices.Map(deployment.Artefacts, ftlv1.ArtefactToProto),
	}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentName)
	if err != nil {
		return err
	}
	defer deployment.Close()

	logger := s.getDeploymentLogger(ctx, deployment.Name)
	logger.Infof("Get deployment artefacts for: %s", deployment.Name)

	chunk := make([]byte, s.config.ArtefactChunkSize)
nextArtefact:
	for _, artefact := range deployment.Artefacts {
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(ftlv1.ArtefactToProto(artefact), clientArtefact) {
				continue nextArtefact
			}
		}
		for {
			n, err := artefact.Content.Read(chunk)
			if n != 0 {
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: ftlv1.ArtefactToProto(artefact),
					Chunk:    chunk[:n],
				}); err != nil {
					return errors.Wrap(err, "could not send artefact chunk")
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return errors.Wrap(err, "could not read artefact chunk")
			}
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	start := time.Now()
	if req.Msg.Verb == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("verb is required"))
	}
	if req.Msg.Body == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("body is required"))
	}
	verbRef := schema.VerbRefFromProto(req.Msg.Verb)

	module := verbRef.Module
	s.routesMu.RLock()
	routes, ok := s.routes[module]
	s.routesMu.RUnlock()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.Errorf("no routes for module %q", module))
	}
	route := routes[rand.Intn(len(routes))] //nolint:gosec
	client := s.clientsForEndpoint(route.Endpoint)

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	requestName, ok, err := headers.GetRequestName(req.Header())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !ok {
		// Inject the request key if this is an ingress call.
		requestName, err = s.dal.CreateIngressRequest(ctx, "grpc", req.Peer().Addr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		headers.SetRequestName(req.Header(), requestName)
	}

	ctx = rpc.WithVerbs(ctx, append(callers, verbRef))
	headers.AddCaller(req.Header(), schema.VerbRefFromProto(req.Msg.Verb))

	resp, err := client.verb.Call(ctx, req)
	var maybeResponse types.Option[*ftlv1.CallResponse]
	if resp != nil {
		maybeResponse = types.Some(resp.Msg)
	}
	s.recordCall(ctx, &Call{
		deploymentName: route.Deployment,
		requestName:    requestName,
		startTime:      start,
		destVerb:       verbRef,
		callers:        callers,
		callError:      types.Nil(err),
		request:        req.Msg,
		response:       maybeResponse,
	})
	return resp, errors.WithStack(err)
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := slices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	need, err := s.dal.GetMissingArtefacts(ctx, byteDigests)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: slices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	logger := log.FromContext(ctx)
	digest, err := s.dal.CreateArtefact(ctx, req.Msg.Content)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logger.Infof("Created new artefact %s", digest)
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: digest[:]}), nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)

	artefacts := make([]dal.DeploymentArtefact, len(req.Msg.Artefacts))
	for i, artefact := range req.Msg.Artefacts {
		digest, err := sha256.ParseSHA256(artefact.Digest)
		if err != nil {
			logger.Errorf(err, "Invalid digest %s", artefact.Digest)
			return nil, errors.Wrap(err, "invalid digest")
		}
		artefacts[i] = dal.DeploymentArtefact{
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
		return nil, errors.Wrap(err, "invalid module schema")
	}
	ingressRoutes := extractIngressRoutingEntries(req.Msg)
	dname, err := s.dal.CreateDeployment(ctx, ms.Runtime.Language, module, artefacts, ingressRoutes)
	if err != nil {
		logger.Errorf(err, "Could not create deployment")
		return nil, errors.Wrap(err, "could not create deployment")
	}
	deploymentLogger := s.getDeploymentLogger(ctx, dname)
	deploymentLogger.Infof("Created deployment %s", dname)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentName: dname.String()}), nil
}

func (s *Service) getDeployment(ctx context.Context, name string) (*model.Deployment, error) {
	dkey, err := model.ParseDeploymentName(name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment name"))
	}
	deployment, err := s.dal.GetDeployment(ctx, dkey)
	if errors.Is(err, pgx.ErrNoRows) {
		logger := s.getDeploymentLogger(ctx, dkey)
		logger.Errorf(err, "Deployment not found")
		return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "could not retrieve deployment"))
	}
	return deployment, nil
}

// Return or create the RunnerService and VerbService clients for a Runner endpoint.
func (s *Service) clientsForEndpoint(endpoint string) clients {
	clientItem := s.clients.Get(endpoint)
	if clientItem != nil {
		return clientItem.Value()
	}
	client := clients{
		runner: rpc.Dial(ftlv1connect.NewRunnerServiceClient, endpoint, log.Error),
		verb:   rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Error),
	}
	s.clients.Set(endpoint, client, time.Minute)
	return client
}

func (s *Service) reapStaleRunners(ctx context.Context) error {
	logger := log.FromContext(ctx)
	count, err := s.dal.KillStaleRunners(context.Background(), s.config.RunnerTimeout)
	if err != nil {
		return errors.Wrap(err, "Failed to delete stale runners")
	} else if count > 0 {
		logger.Warnf("Reaped %d stale runners", count)
	}
	return nil
}

func (s *Service) releaseExpiredReservations(ctx context.Context) error {
	logger := log.FromContext(ctx)
	count, err := s.dal.ExpireRunnerClaims(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to expire runner reservations")
	} else if count > 0 {
		logger.Warnf("Expired %d runner reservations", count)
	}
	return nil
}

func (s *Service) reconcileDeployments(ctx context.Context) error {
	reconciliation, err := s.dal.GetDeploymentsNeedingReconciliation(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get deployments needing reconciliation")
	}
	wg, ctx := concurrency.New(ctx, concurrency.WithConcurrencyLimit(4))
	for _, reconcile := range reconciliation {
		reconcile := reconcile
		deploymentLogger := s.getDeploymentLogger(ctx, reconcile.Deployment)
		deploymentLogger.Infof("Reconciling %s", reconcile.Deployment)
		deployment := model.Deployment{
			Module:   reconcile.Module,
			Language: reconcile.Language,
			Name:     reconcile.Deployment,
		}
		require := reconcile.RequiredReplicas - reconcile.AssignedReplicas
		if require > 0 {
			deploymentLogger.Infof("Need %d more runners for %s", require, reconcile.Deployment)
			wg.Go(func(ctx context.Context) error {
				if err := s.deploy(ctx, deployment); err != nil {
					deploymentLogger.Warnf("Failed to increase deployment replicas: %s", err)
				} else {
					deploymentLogger.Infof("Reconciled %s", reconcile.Deployment)
				}
				return nil
			})
		} else if require < 0 {
			deploymentLogger.Infof("Need %d less runners for %s", -require, reconcile.Deployment)
			wg.Go(func(ctx context.Context) error {
				ok, err := s.terminateRandomRunner(ctx, deployment.Name)
				if err != nil {
					deploymentLogger.Warnf("Failed to terminate runner: %s", err)
				} else if ok {
					deploymentLogger.Infof("Reconciled %s", reconcile.Deployment)
				} else {
					deploymentLogger.Warnf("Failed to terminate runner: no runners found")
				}
				return nil
			})
		}
	}
	_ = wg.Wait()
	return nil
}

func (s *Service) reconcileRunners(ctx context.Context) error {
	activeDeployments, err := s.dal.GetActiveDeployments(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get deployments needing reconciliation")
	}

	totalRunners := 0
	for _, deployment := range activeDeployments {
		totalRunners += deployment.MinReplicas
	}

	// It's possible that idles runners will get terminated here, but they will get recreated in the next
	// reconciliation cycle.
	idleRunners, err := s.dal.GetIdleRunners(ctx, 16, model.Labels{})
	if err != nil {
		return errors.WithStack(err)
	}

	idleRunnerKeys := slices.Map(idleRunners, func(r dal.Runner) model.RunnerKey { return r.Key })

	err = s.runnerScaling.SetReplicas(ctx, totalRunners, idleRunnerKeys)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *Service) terminateRandomRunner(ctx context.Context, key model.DeploymentName) (bool, error) {
	runners, err := s.dal.GetRunnersForDeployment(ctx, key)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get runner for %s", key)
	}
	if len(runners) == 0 {
		return false, nil
	}
	runner := runners[rand.Intn(len(runners))] //nolint:gosec
	client := s.clientsForEndpoint(runner.Endpoint)
	resp, err := client.runner.Terminate(ctx, connect.NewRequest(&ftlv1.TerminateRequest{DeploymentName: key.String()}))
	if err != nil {
		return false, errors.WithStack(err)
	}
	err = s.dal.UpsertRunner(ctx, dal.Runner{
		Key:      runner.Key,
		Endpoint: runner.Endpoint,
		State:    dal.RunnerStateFromProto(resp.Msg.State),
		Labels:   runner.Labels,
	})
	return true, errors.WithStack(err)
}

func (s *Service) deploy(ctx context.Context, reconcile model.Deployment) error {
	client, err := s.reserveRunner(ctx, reconcile)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = client.runner.Deploy(ctx, connect.NewRequest(&ftlv1.DeployRequest{DeploymentName: reconcile.Name.String()}))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Service) reserveRunner(ctx context.Context, reconcile model.Deployment) (client clients, err error) {
	// A timeout context applied to the transaction and the Runner.Reserve() Call.
	reservationCtx, cancel := context.WithTimeout(ctx, s.config.DeploymentReservationTimeout)
	defer cancel()
	claim, err := s.dal.ReserveRunnerForDeployment(reservationCtx, reconcile.Name, s.config.DeploymentReservationTimeout, model.Labels{
		"languages": []string{reconcile.Language},
	})
	if err != nil {
		return clients{}, errors.Wrapf(err, "failed to claim runners for %s", reconcile.Name)
	}

	err = errors.WithStack(dal.WithReservation(reservationCtx, claim, func() error {
		client = s.clientsForEndpoint(claim.Runner().Endpoint)
		_, err = client.runner.Reserve(reservationCtx, connect.NewRequest(&ftlv1.ReserveRequest{DeploymentName: reconcile.Name.String()}))
		return errors.WithStack(err)
	}))
	return
}

// Periodically remove stale (ie. have not heartbeat recently) controllers from the database.
func (s *Service) reapStaleControllers(ctx context.Context) error {
	logger := log.FromContext(ctx)
	count, err := s.dal.KillStaleControllers(context.Background(), s.config.RunnerTimeout)
	if err != nil {
		return errors.Wrap(err, "failed to delete stale controllers")
	} else if count > 0 {
		logger.Warnf("Reaped %d stale controllers", count)
	}
	return nil
}

// Periodically update the DB with the current state of the controller.
func (s *Service) heartbeatController(ctx context.Context) error {
	_, err := s.dal.UpsertController(ctx, s.key, s.config.Advertise.String())
	if err != nil {
		return errors.Wrap(err, "failed to heartbeat controller")
	}
	return nil

}

func (s *Service) watchModuleChanges(ctx context.Context, sendChange func(response *ftlv1.PullSchemaResponse) error) error {
	logger := log.FromContext(ctx)
	type moduleStateEntry struct {
		hash        sha256.SHA256
		minReplicas int
	}
	moduleState := map[string]moduleStateEntry{}
	moduleByDeploymentName := map[model.DeploymentName]string{}

	// Seed the notification channel with the current deployments.
	seedDeployments, err := s.dal.GetActiveDeployments(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	initialCount := len(seedDeployments)
	deploymentChanges := make(chan dal.DeploymentNotification, len(seedDeployments))
	for _, deployment := range seedDeployments {
		deploymentChanges <- dal.DeploymentNotification{Message: types.Some(deployment)}
	}
	logger.Infof("Seeded %d deployments", initialCount)

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
				name := moduleByDeploymentName[deletion]
				response = &ftlv1.PullSchemaResponse{
					ModuleName:     name,
					DeploymentName: deletion.String(),
					ChangeType:     ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED,
				}
				delete(moduleState, name)
				delete(moduleByDeploymentName, deletion)
			} else if message, ok := notification.Message.Get(); ok {
				moduleSchema := message.Schema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
				moduleSchema.Runtime = &schemapb.ModuleRuntime{
					Language:    message.Language,
					CreateTime:  timestamppb.New(message.CreatedAt),
					MinReplicas: int32(message.MinReplicas),
				}
				moduleSchemaBytes, err := proto.Marshal(moduleSchema)
				if err != nil {
					return errors.WithStack(err)
				}
				newState := moduleStateEntry{
					hash:        sha256.FromBytes(moduleSchemaBytes),
					minReplicas: message.MinReplicas,
				}
				if current, ok := moduleState[message.Schema.Name]; ok {
					if current != newState {
						changeType := ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED
						// A deployment is considered removed if its minReplicas is set to 0.
						if current.minReplicas > 0 && message.MinReplicas == 0 {
							changeType = ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED
						}
						response = &ftlv1.PullSchemaResponse{
							ModuleName:     moduleSchema.Name,
							DeploymentName: message.Name.String(),
							Schema:         moduleSchema,
							ChangeType:     changeType,
						}
					}
				} else {
					response = &ftlv1.PullSchemaResponse{
						ModuleName:     moduleSchema.Name,
						DeploymentName: message.Name.String(),
						Schema:         moduleSchema,
						ChangeType:     ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
						More:           initialCount > 1,
					}
					if initialCount > 0 {
						initialCount--
					}
				}
				moduleState[message.Schema.Name] = newState
				delete(moduleByDeploymentName, message.Name) // The deployment may have changed.
				moduleByDeploymentName[message.Name] = message.Schema.Name
			}

			if response != nil {
				logger.Tracef("Sending change %s", response.ChangeType)
				err := sendChange(response)
				if err != nil {
					return errors.WithStack(err)
				}
			} else {
				logger.Tracef("No change")
			}
		}
	}
}

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentName model.DeploymentName) *log.Logger {
	attrs := map[string]string{"deployment": deploymentName.String()}
	if requestName, ok, _ := rpc.RequestNameFromContext(ctx); ok {
		attrs["request"] = requestName.String()
	}

	return log.FromContext(ctx).AddSink(s.deploymentLogsSink).Sub(attrs)
}

// Periodically sync the routing table from the DB.
func (s *Service) syncRoutes(ctx context.Context) error {
	routes, err := s.dal.GetRoutingTable(ctx, nil)
	if errors.Is(err, dal.ErrNotFound) {
		routes = map[string][]dal.Route{}
	} else if err != nil {
		return errors.WithStack(err)
	}
	s.routesMu.Lock()
	s.routes = routes
	s.routesMu.Unlock()
	return nil
}

// Copied from the Apache-licensed connect-go source.
func connectCodeToHTTP(code connect.Code) int {
	switch code {
	case connect.CodeCanceled:
		return 408
	case connect.CodeUnknown:
		return 500
	case connect.CodeInvalidArgument:
		return 400
	case connect.CodeDeadlineExceeded:
		return 408
	case connect.CodeNotFound:
		return 404
	case connect.CodeAlreadyExists:
		return 409
	case connect.CodePermissionDenied:
		return 403
	case connect.CodeResourceExhausted:
		return 429
	case connect.CodeFailedPrecondition:
		return 412
	case connect.CodeAborted:
		return 409
	case connect.CodeOutOfRange:
		return 400
	case connect.CodeUnimplemented:
		return 404
	case connect.CodeInternal:
		return 500
	case connect.CodeUnavailable:
		return 503
	case connect.CodeDataLoss:
		return 500
	case connect.CodeUnauthenticated:
		return 401
	default:
		return 500 // same as CodeUnknown
	}
}

func runWithRetries(ctx context.Context, success, failure time.Duration, fn func(ctx context.Context) error) {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	name = name[strings.LastIndex(name, ".")+1:]
	name = strings.TrimSuffix(name, "-fm")

	ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope(name))
	failureRetry := backoff.Backoff{
		Min:    failure,
		Max:    failure * 2,
		Jitter: true,
		Factor: 2,
	}
	failed := false
	logger := log.FromContext(ctx)
	for {
		err := fn(ctx)
		if err != nil {
			next := failureRetry.Duration()
			logger.Errorf(err, "Failed, retrying in %s", next)
			select {
			case <-time.After(next):
			case <-ctx.Done():
				return
			}
		} else {
			if failed {
				logger.Infof("Recovered")
				failed = false
			}
			failureRetry.Reset()
			logger.Tracef("Success, next run in %s", success)
			select {
			case <-time.After(success):
			case <-ctx.Done():
				return
			}
		}
	}
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
						Path:   ingress.Ingress.Path,
					})
				}
			}
		}
	}
	return ingressRoutes
}
