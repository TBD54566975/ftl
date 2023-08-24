package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
	"github.com/bufbuild/connect-go"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpillora/backoff"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/backend/common/rpc/headers"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/controller/internal/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/console"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Config struct {
	Bind                         *url.URL            `help:"Socket to bind to." default:"http://localhost:8892" env:"FTL_CONTROLLER_BIND"`
	Advertise                    *url.URL            `help:"Endpoint the Controller should advertise (use --bind if omitted)." default:"" env:"FTL_CONTROLLER_ADVERTISE"`
	Key                          model.ControllerKey `help:"Controller key (auto)." placeholder:"C<ULID>" default:"C00000000000000000000000000"`
	DSN                          string              `help:"DAL DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_CONTROLLER_DSN"`
	RunnerTimeout                time.Duration       `help:"Runner heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration       `help:"Deployment reservation timeout." default:"120s"`
	ArtefactChunkSize            int                 `help:"Size of each chunk streamed to the client." default:"1048576"`
}

// Start the Controller. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL controller")

	c, err := console.Server(ctx)
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

	svc, err := New(ctx, dal, config)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Listening on %s", config.Bind)

	console := NewConsoleService(dal)

	return rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc.GRPC(ftlv1connect.NewControllerServiceHandler, svc),
		rpc.GRPC(pbconsoleconnect.NewConsoleServiceHandler, console),
		rpc.HTTP("/ingress/", http.StripPrefix("/ingress", svc)),
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
	dal                          *dal.DAL
	heartbeatTimeout             time.Duration
	deploymentReservationTimeout time.Duration
	artefactChunkSize            int
	key                          model.ControllerKey
	deploymentLogsSink           *deploymentLogsSink

	// Map from endpoint to client.
	clients *lru.Cache[string, clients]
}

func New(ctx context.Context, dal *dal.DAL, config Config) (*Service, error) {
	key := config.Key
	if config.Key.ULID() == (ulid.ULID{}) {
		key = model.NewControllerKey()
	}
	clientCache, err := lru.New[string, clients](1024)
	if err != nil {
		return nil, errors.Wrap(err, "could not create client cache")
	}
	svc := &Service{
		dal:                          dal,
		heartbeatTimeout:             config.RunnerTimeout,
		deploymentReservationTimeout: config.DeploymentReservationTimeout,
		artefactChunkSize:            config.ArtefactChunkSize,
		key:                          key,
		deploymentLogsSink:           newDeploymentLogsSink(ctx, dal),
		clients:                      clientCache,
	}
	if config.Advertise.String() == "" {
		config.Advertise = config.Bind
	}
	go svc.heartbeatController(ctx, config.Advertise)
	go svc.reapStaleControllers(ctx)
	go svc.reapStaleRunners(ctx)
	go svc.releaseExpiredReservations(ctx)
	go svc.reconcileDeployments(ctx)
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
	client := s.clientsForEndpoint(route.Endpoint)
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
	resp, err := client.verb.Call(r.Context(), connect.NewRequest(&ftlv1.CallRequest{
		Verb: &pschema.VerbRef{Module: route.Module, Name: route.Verb},
		Body: body,
	}))
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

func (s *Service) Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	status, err := s.dal.GetStatus(ctx, req.Msg.AllControllers, req.Msg.AllRunners, req.Msg.AllDeployments, req.Msg.AllIngressRoutes)
	if err != nil {
		return nil, errors.Wrap(err, "could not get status")
	}
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
			Schema:      d.Schema.ToProto().(*pschema.Module), //nolint:forcetypeassert
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
				Module:         r.Module,
				Verb:           r.Verb,
				Method:         r.Method,
				Path:           r.Path,
			}
		}),
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
		var requestKey types.Option[model.IngressRequestKey]
		if msg.RequestKey != nil {
			rkey, err := model.ParseIngressRequestKey(*msg.RequestKey)
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid request key"))
			}
			requestKey = types.Some(rkey)
		}

		err = s.dal.InsertLogEvent(ctx, &dal.LogEvent{
			RequestKey:     requestKey,
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
	sch := &pschema.Schema{
		Modules: slices.Map(deployments, func(d dal.Deployment) *pschema.Module {
			return d.Schema.ToProto().(*pschema.Module) //nolint:forcetypeassert
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
			logger.Errorf(err, "Deployment already exists: %s", newDeploymentName)
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.WithStack(err))
		}
		logger.Errorf(err, "Could not replace deployment: %s", newDeploymentName)
		return nil, errors.Wrap(err, "could not replace deployment")
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
	heartbeatCtx, cancel := context.WithTimeout(ctx, s.heartbeatTimeout)
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
		Schema:    deployment.Schema.ToProto().(*pschema.Module), //nolint:forcetypeassert
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

	chunk := make([]byte, s.artefactChunkSize)
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

	routes, err := s.dal.GetRoutingTable(ctx, req.Msg.Verb.Module)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "no runners for module %q", req.Msg.Verb.Module))
		}
		return nil, errors.Wrap(err, "failed to get runners for module")
	}
	route := routes[rand.Intn(len(routes))] //nolint:gosec
	client := s.clientsForEndpoint(route.Endpoint)

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var requestKey model.IngressRequestKey
	if len(callers) == 0 {
		// Inject the request key if this is an ingress call.
		requestKey, err = s.dal.CreateIngressRequest(ctx, req.Peer().Addr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		headers.SetRequestKey(req.Header(), requestKey)
	} else {
		var ok bool
		requestKey, ok, err = headers.GetRequestKey(req.Header())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if !ok {
			return nil, errors.New("request Key is missing")
		}
	}

	callRecord := &Call{
		deploymentName: route.Deployment,
		requestKey:     requestKey,
		startTime:      start,
		destVerb:       verbRef,
		callers:        callers,
		request:        req.Msg,
	}

	ctx = rpc.WithVerbs(ctx, append(callers, verbRef))
	headers.AddCaller(req.Header(), schema.VerbRefFromProto(req.Msg.Verb))

	resp, err := client.verb.Call(ctx, req)
	if err != nil {
		s.recordCallError(ctx, callRecord, err)
		return nil, errors.WithStack(err)
	}

	callRecord.response = resp.Msg
	s.recordCall(ctx, callRecord)
	return resp, nil
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
	client, ok := s.clients.Get(endpoint)
	if ok {
		return client
	}
	client = clients{
		runner: rpc.Dial(ftlv1connect.NewRunnerServiceClient, endpoint, log.Error),
		verb:   rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Error),
	}
	s.clients.Add(endpoint, client)
	return client
}

func (s *Service) reapStaleRunners(ctx context.Context) {
	logger := log.FromContext(ctx)
	for {
		count, err := s.dal.KillStaleRunners(context.Background(), s.heartbeatTimeout)
		if err != nil {
			logger.Errorf(err, "Failed to delete stale runners")
		} else if count > 0 {
			logger.Warnf("Reaped %d stale runners", count)
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(s.heartbeatTimeout / 4):
		}
	}
}

func (s *Service) releaseExpiredReservations(ctx context.Context) {
	logger := log.FromContext(ctx)
	for {
		count, err := s.dal.ExpireRunnerClaims(ctx)
		if err != nil {
			logger.Errorf(err, "Failed to expire runner reservations")
		} else if count > 0 {
			logger.Warnf("Expired %d runner reservations", count)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.deploymentReservationTimeout):
		}
	}
}

func (s *Service) reconcileDeployments(ctx context.Context) {
	logger := log.FromContext(ctx)
	for {
		reconciliation, err := s.dal.GetDeploymentsNeedingReconciliation(ctx)
		if err != nil {
			logger.Errorf(err, "Failed to get deployments needing reconciliation")
		} else {
			wg, ctx := concurrency.New(ctx, concurrency.WithConcurrencyLimit(4)) //nolint:govet
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
		}

		select {
		case <-ctx.Done():
			return

		case <-time.After(time.Second):
		}
	}
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
	reservationCtx, cancel := context.WithTimeout(ctx, s.deploymentReservationTimeout)
	defer cancel()
	claim, err := s.dal.ReserveRunnerForDeployment(reservationCtx, reconcile.Name, s.deploymentReservationTimeout, model.Labels{
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

func (s *Service) reapStaleControllers(ctx context.Context) {
	logger := log.FromContext(ctx)
	for {
		count, err := s.dal.KillStaleControllers(context.Background(), s.heartbeatTimeout)
		if err != nil {
			logger.Errorf(err, "Failed to delete stale controllers")
		} else if count > 0 {
			logger.Warnf("Reaped %d stale controllers", count)
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(s.heartbeatTimeout / 4):
		}
	}
}

// Periodically update the DB with the current state of the controller.
func (s *Service) heartbeatController(ctx context.Context, advertiseAddr *url.URL) {
	logger := log.FromContext(ctx)
	for {
		_, err := s.dal.UpsertController(ctx, s.key, advertiseAddr.String())
		if err != nil {
			logger.Errorf(err, "Failed to heartbeat controller")
		}
		select {
		case <-ctx.Done():
			return

		case <-time.After(s.heartbeatTimeout / 4):
		}
	}
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
		deploymentChanges <- dal.DeploymentNotification{Message: deployment}
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
			if key, ok := notification.Deleted.Get(); ok {
				name := moduleByDeploymentName[key]
				response = &ftlv1.PullSchemaResponse{
					ModuleName:     name,
					DeploymentName: key.String(),
					ChangeType:     ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED,
				}
				delete(moduleState, name)
				delete(moduleByDeploymentName, key)
			} else {
				moduleSchema := notification.Message.Schema.ToProto().(*pschema.Module) //nolint:forcetypeassert
				moduleSchema.Runtime = &pschema.ModuleRuntime{
					Language:    notification.Message.Language,
					CreateTime:  timestamppb.New(notification.Message.CreatedAt),
					MinReplicas: int32(notification.Message.MinReplicas),
				}
				moduleSchemaBytes, err := proto.Marshal(moduleSchema)
				if err != nil {
					return errors.WithStack(err)
				}
				newState := moduleStateEntry{
					hash:        sha256.FromBytes(moduleSchemaBytes),
					minReplicas: notification.Message.MinReplicas,
				}
				if current, ok := moduleState[notification.Message.Schema.Name]; ok {
					if current != newState {
						changeType := ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED
						// A deployment is considered removed if its minReplicas is set to 0.
						if current.minReplicas > 0 && notification.Message.MinReplicas == 0 {
							changeType = ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED
						}
						response = &ftlv1.PullSchemaResponse{
							ModuleName:     moduleSchema.Name,
							DeploymentName: notification.Message.Name.String(),
							Schema:         moduleSchema,
							ChangeType:     changeType,
						}
					}
				} else {
					response = &ftlv1.PullSchemaResponse{
						ModuleName:     moduleSchema.Name,
						DeploymentName: notification.Message.Name.String(),
						Schema:         moduleSchema,
						ChangeType:     ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
						More:           initialCount > 1,
					}
					if initialCount > 0 {
						initialCount--
					}
				}
				moduleState[notification.Message.Schema.Name] = newState
				delete(moduleByDeploymentName, notification.Message.Name) // The deployment may have changed.
				moduleByDeploymentName[notification.Message.Name] = notification.Message.Schema.Name
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

func extractIngressRoutingEntries(req *ftlv1.CreateDeploymentRequest) []dal.IngressRoutingEntry {
	var ingressRoutes []dal.IngressRoutingEntry
	for _, decl := range req.Schema.Decls {
		if verb, ok := decl.Value.(*pschema.Decl_Verb); ok {
			for _, metadata := range verb.Verb.Metadata {
				if ingress, ok := metadata.Value.(*pschema.Metadata_Ingress); ok {
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

func (s *Service) getDeploymentLogger(ctx context.Context, deploymentName model.DeploymentName) *log.Logger {
	attrs := map[string]string{"deployment": deploymentName.String()}
	if requestKey, ok, _ := rpc.RequestKeyFromContext(ctx); ok {
		attrs["request"] = requestKey.String()
	}

	return log.FromContext(ctx).AddSink(s.deploymentLogsSink).Sub(attrs)
}
