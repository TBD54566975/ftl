package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpillora/backoff"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/console"
	"github.com/TBD54566975/ftl/controller/internal/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type Config struct {
	Bind                         *url.URL            `help:"Socket to bind to." default:"http://localhost:8892"`
	Key                          model.ControllerKey `help:"Controller key (auto)." placeholder:"C<ULID>" default:"C00000000000000000000000000"`
	DSN                          string              `help:"DAL DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret"`
	RunnerTimeout                time.Duration       `help:"Runner heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration       `help:"Deployment reservation timeout." default:"120s"`
	ArtefactChunkSize            int                 `help:"Size of each chunk streamed to the client." default:"1048576"`
}

// Start the Controller. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL controller")
	conn, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil
	}

	c, err := console.Server(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	dal := dal.New(conn)
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

	clientsMu sync.Mutex
	// Map from endpoint to client.
	clients map[string]clients
}

func New(ctx context.Context, dal *dal.DAL, config Config) (*Service, error) {
	key := config.Key
	if config.Key.ULID() == (ulid.ULID{}) {
		key = model.NewControllerKey()
	}
	svc := &Service{
		dal:                          dal,
		heartbeatTimeout:             config.RunnerTimeout,
		deploymentReservationTimeout: config.DeploymentReservationTimeout,
		artefactChunkSize:            config.ArtefactChunkSize,
		clients:                      map[string]clients{},
		key:                          key,
	}
	go svc.heartbeatController(ctx, config.Bind)
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
	resp := &ftlv1.StatusResponse{
		Controllers: slices.Map(status.Controllers, func(c dal.Controller) *ftlv1.StatusResponse_Controller {
			return &ftlv1.StatusResponse_Controller{
				Key:      c.Key.String(),
				Endpoint: c.Endpoint,
				State:    c.State.ToProto(),
			}
		}),
		Runners: slices.Map(status.Runners, func(r dal.Runner) *ftlv1.StatusResponse_Runner {
			var deployment *string
			if d, ok := r.Deployment.Get(); ok {
				asString := d.String()
				deployment = &asString
			}
			return &ftlv1.StatusResponse_Runner{
				Key:        r.Key.String(),
				Language:   r.Language,
				Endpoint:   r.Endpoint,
				State:      r.State.ToProto(),
				Deployment: deployment,
			}
		}),
		Deployments: slices.Map(status.Deployments, func(d dal.Deployment) *ftlv1.StatusResponse_Deployment {
			return &ftlv1.StatusResponse_Deployment{
				Key:         d.Key.String(),
				Language:    d.Language,
				Name:        d.Module,
				MinReplicas: int32(d.MinReplicas),
				Schema:      d.Schema.ToProto().(*pschema.Module), //nolint:forcetypeassert
			}
		}),
		IngressRoutes: slices.Map(status.IngressRoutes, func(r dal.IngressRouteEntry) *ftlv1.StatusResponse_IngressRoute {
			return &ftlv1.StatusResponse_IngressRoute{
				DeploymentKey: r.Deployment.String(),
				Module:        r.Module,
				Verb:          r.Verb,
				Method:        r.Method,
				Path:          r.Path,
			}
		}),
	}
	return connect.NewResponse(resp), nil
}

func (s *Service) StreamDeploymentLogs(ctx context.Context, req *connect.ClientStream[ftlv1.StreamDeploymentLogsRequest]) (*connect.Response[ftlv1.StreamDeploymentLogsResponse], error) {
	panic("unimplemented")
}

func (s *Service) PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest], stream *connect.ServerStream[ftlv1.PullSchemaResponse]) error {
	return s.watchModuleChanges(ctx, func(response *ftlv1.PullSchemaResponse) error {
		return errors.WithStack(stream.Send(response))
	})
}

func (s *Service) UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (response *connect.Response[ftlv1.UpdateDeployResponse], err error) {
	deploymentKey, err := model.ParseDeploymentKey(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	err = s.dal.SetDeploymentReplicas(ctx, deploymentKey, int(req.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		}
		return nil, errors.Wrap(err, "could not set deployment replicas")
	}

	return connect.NewResponse(&ftlv1.UpdateDeployResponse{}), nil
}

func (s *Service) ReplaceDeploy(ctx context.Context, c *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	newDeploymentKey, err := model.ParseDeploymentKey(c.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
	}
	err = s.dal.ReplaceDeployment(ctx, newDeploymentKey, int(c.Msg.MinReplicas))
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
		} else if errors.Is(err, dal.ErrConflict) {
			return nil, connect.NewError(connect.CodeAlreadyExists, errors.WithStack(err))
		}
		return nil, errors.Wrap(err, "could not replace deployment")
	}
	return connect.NewResponse(&ftlv1.ReplaceDeployResponse{}), nil
}

func (s *Service) RegisterRunner(ctx context.Context, stream *connect.ClientStream[ftlv1.RunnerHeartbeat]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {
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
			Language:   msg.Language,
			Endpoint:   msg.Endpoint,
			State:      dal.RunnerStateFromProto(msg.State),
			Deployment: maybeDeployment,
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
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema:    deployment.Schema.ToProto().(*pschema.Module), //nolint:forcetypeassert
		Artefacts: slices.Map(deployment.Artefacts, ftlv1.ArtefactToProto),
	}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return err
	}
	defer deployment.Close()
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

	callRecord := &call{
		requestKey:    requestKey,
		controllerKey: s.key,
		runnerKey:     route.Runner,
		startTime:     start,
		destVerb:      verbRef,
		callers:       callers,
		request:       req.Msg,
	}

	ctx = rpc.WithVerbs(ctx, append(callers, verbRef))
	headers.AddCaller(req.Header(), schema.VerbRefFromProto(req.Msg.Verb))

	resp, err := client.verb.Call(ctx, req)
	if err != nil {
		s.recordCallError(ctx, callRecord, err)
		return nil, errors.WithStack(err)
	}

<<<<<<< HEAD
	err = s.recordCall(ctx, &Call{
		requestKey:    requestKey,
		runnerKey:     route.Runner,
		controllerKey: s.key,
		startTime:     start,
		destVerb:      verbRef,
		callers:       callers,
		request:       req.Msg,
		response:      resp.Msg,
	})

=======
	callRecord.response = resp.Msg
	err = s.recordCall(ctx, callRecord)
>>>>>>> b71a577 (Add more error handling to ftl calls)
	if err != nil {
		return nil, errors.WithStack(err)
	}
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
		return nil, errors.New("missing runtime metadata")
	}
	module, err := schema.ModuleFromProto(ms)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module schema")
	}
	ingressRoutes := extractIngressRoutingEntries(req.Msg)
	key, err := s.dal.CreateDeployment(ctx, ms.Runtime.Language, module, artefacts, ingressRoutes)
	if err != nil {
		return nil, errors.Wrap(err, "could not create deployment")
	}
	logger.Infof("Created deployment %s", key)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: key.String()}), nil
}

func (s *Service) getDeployment(ctx context.Context, key string) (*model.Deployment, error) {
	dkey, err := model.ParseDeploymentKey(key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}
	deployment, err := s.dal.GetDeployment(ctx, dkey)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("deployment not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, errors.Wrap(err, "could not retrieve deployment"))
	}
	return deployment, nil
}

// Return or create the RunnerService and VerbService clients for a Runner endpoint.
func (s *Service) clientsForEndpoint(endpoint string) clients {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	client, ok := s.clients[endpoint]
	if ok {
		return client
	}
	client = clients{
		runner: rpc.Dial(ftlv1connect.NewRunnerServiceClient, endpoint, log.Error),
		verb:   rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Error),
	}
	s.clients[endpoint] = client
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
			for _, reconcile := range reconciliation {
				logger.Infof("Reconciling %s", reconcile.Deployment)
				deployment := model.Deployment{
					Module:   reconcile.Module,
					Language: reconcile.Language,
					Key:      reconcile.Deployment,
				}
				require := reconcile.RequiredReplicas - reconcile.AssignedReplicas
				if require > 0 {
					logger.Infof("Need %d more runners for %s", require, reconcile.Deployment)
					if err := s.deploy(ctx, deployment); err != nil {
						logger.Warnf("Failed to increase deployment replicas: %s", err)
					} else {
						logger.Infof("Reconciled %s", reconcile.Deployment)
					}
				} else if require < 0 {
					logger.Infof("Need %d less runners for %s", -require, reconcile.Deployment)
					ok, err := s.terminateRandomRunner(ctx, deployment.Key)
					if err != nil {
						logger.Warnf("Failed to terminate runner: %s", err)
					} else if ok {
						logger.Infof("Reconciled %s", reconcile.Deployment)
					} else {
						logger.Warnf("Failed to terminate runner: no runners found")
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			return

		case <-time.After(time.Second):
		}
	}
}

func (s *Service) terminateRandomRunner(ctx context.Context, key model.DeploymentKey) (bool, error) {
	runners, err := s.dal.GetRunnersForDeployment(ctx, key)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get runner for %s", key)
	}
	if len(runners) == 0 {
		return false, nil
	}
	runner := runners[rand.Intn(len(runners))] //nolint:gosec
	client := s.clientsForEndpoint(runner.Endpoint)
	resp, err := client.runner.Terminate(ctx, connect.NewRequest(&ftlv1.TerminateRequest{DeploymentKey: key.String()}))
	if err != nil {
		return false, errors.WithStack(err)
	}
	err = s.dal.UpsertRunner(ctx, dal.Runner{
		Key:      runner.Key,
		Language: runner.Language,
		Endpoint: runner.Endpoint,
		State:    dal.RunnerStateFromProto(resp.Msg.State),
	})
	return true, errors.WithStack(err)
}

func (s *Service) deploy(ctx context.Context, reconcile model.Deployment) error {
	client, err := s.reserveRunner(ctx, reconcile)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = client.runner.Deploy(ctx, connect.NewRequest(&ftlv1.DeployRequest{DeploymentKey: reconcile.Key.String()}))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *Service) reserveRunner(ctx context.Context, reconcile model.Deployment) (client clients, err error) {
	// A timeout context applied to the transaction and the Runner.Reserve() Call.
	reservationCtx, cancel := context.WithTimeout(ctx, s.deploymentReservationTimeout)
	defer cancel()
	claim, err := s.dal.ReserveRunnerForDeployment(reservationCtx, reconcile.Language, reconcile.Key, s.deploymentReservationTimeout)
	if err != nil {
		return clients{}, errors.Wrapf(err, "failed to claim runners for %s", reconcile.Key)
	}

	err = errors.WithStack(dal.WithReservation(reservationCtx, claim, func() error {
		client = s.clientsForEndpoint(claim.Runner().Endpoint)
		_, err = client.runner.Reserve(reservationCtx, connect.NewRequest(&ftlv1.ReserveRequest{DeploymentKey: reconcile.Key.String()}))
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
func (s *Service) heartbeatController(ctx context.Context, addr *url.URL) {
	logger := log.FromContext(ctx)
	for {
		_, err := s.dal.UpsertController(ctx, s.key, addr.String())
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
	moduleSchemas := map[string]schema.Module{}

	for {
		dbDeployments := map[string]dal.Deployment{}
		var changesToSend []*ftlv1.PullSchemaResponse

		deployments, err := s.dal.GetActiveDeployments(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		for _, deployment := range deployments {
			dbDeployments[deployment.Schema.Name] = deployment
			if current, ok := moduleSchemas[deployment.Schema.Name]; ok {
				if !proto.Equal(current.ToProto(), deployment.Schema.ToProto()) {
					//nolint:forcetypeassert
					changesToSend = append(changesToSend, &ftlv1.PullSchemaResponse{
						DeploymentKey: deployment.Key.String(),
						Schema:        deployment.Schema.ToProto().(*pschema.Module),
						ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED,
					})
				}
			} else {
				//nolint:forcetypeassert
				changesToSend = append(changesToSend, &ftlv1.PullSchemaResponse{
					DeploymentKey: deployment.Key.String(),
					Schema:        deployment.Schema.ToProto().(*pschema.Module),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED,
				})
			}
			moduleSchemas[deployment.Schema.Name] = *deployment.Schema
		}

		for name, moduleSchema := range moduleSchemas {
			if dbDeployment, ok := dbDeployments[name]; !ok {
				//nolint:forcetypeassert
				changesToSend = append(changesToSend, &ftlv1.PullSchemaResponse{
					DeploymentKey: dbDeployment.Key.String(),
					Schema:        moduleSchema.ToProto().(*pschema.Module),
					ChangeType:    ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED,
				})
				delete(moduleSchemas, name)
			}
		}

		for i, change := range changesToSend {
			err := sendChange(&ftlv1.PullSchemaResponse{
				DeploymentKey: change.DeploymentKey,
				Schema:        change.Schema,
				More:          i < len(changesToSend)-1,
				ChangeType:    change.ChangeType,
			})
			if err != nil {
				return errors.WithStack(err)
			}
		}

		select {
		case <-ctx.Done():
			return nil

		case <-time.After(1 * time.Second):
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
