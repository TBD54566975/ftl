package controlplane

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jpillora/backoff"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/console"
	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type Config struct {
	Bind                         *url.URL      `help:"Socket to bind to." default:"http://localhost:8892"`
	DSN                          string        `help:"Postgres DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret"`
	RunnerTimeout                time.Duration `help:"Runner heartbeat timeout." default:"10s"`
	DeploymentReservationTimeout time.Duration `help:"Deployment reservation timeout." default:"120s"`
	ArtefactChunkSize            int           `help:"Size of each chunk streamed to the client." default:"1048576"`
}

// Start the ControlPlane. Blocks until the context is cancelled.
func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL controlplane")
	conn, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil
	}

	c, err := console.Server(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	svc, err := New(ctx, dal.New(conn), config.RunnerTimeout, config.DeploymentReservationTimeout, config.ArtefactChunkSize)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Listening on %s", config.Bind)

	return rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		rpc.GRPC(ftlv1connect.NewControlPlaneServiceHandler, svc),
		rpc.Route("/", c),
	)
}

var _ ftlv1connect.ControlPlaneServiceHandler = (*Service)(nil)
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

	clientsMu sync.Mutex
	// Map from endpoint to client.
	clients map[string]clients
}

func New(ctx context.Context, dal *dal.DAL, heartbeatTimeout, deploymentReservationTimeout time.Duration, artefactChunkSize int) (*Service, error) {
	svc := &Service{
		dal:                          dal,
		heartbeatTimeout:             heartbeatTimeout,
		deploymentReservationTimeout: deploymentReservationTimeout,
		artefactChunkSize:            artefactChunkSize,
		clients:                      map[string]clients{},
	}
	go svc.reapStaleRunners(ctx)
	go svc.releaseExpiredReservations(ctx)
	return svc, nil
}

func (s *Service) StreamDeploymentLogs(ctx context.Context, req *connect.ClientStream[ftlv1.StreamDeploymentLogsRequest]) (*connect.Response[ftlv1.StreamDeploymentLogsResponse], error) {
	panic("unimplemented")
}

func (s *Service) Deploy(ctx context.Context, req *connect.Request[ftlv1.DeployRequest]) (response *connect.Response[ftlv1.DeployResponse], err error) {
	deploymentKey, err := ulid.Parse(req.Msg.DeploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid deployment key"))
	}

	deployment, err := s.dal.GetDeployment(ctx, deploymentKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not find requested deployment")
	}

	runner, err := s.dal.ReserveRunnerForDeployment(ctx, deployment.Language, deploymentKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeResourceExhausted, errors.Wrap(err, "failed to reserve runner"))
	}

	client := s.clientsForEndpoint(runner.Endpoint)
	_, err = client.runner.DeployToRunner(ctx, connect.NewRequest(&ftlv1.DeployToRunnerRequest{DeploymentKey: deploymentKey.String()}))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&ftlv1.DeployResponse{}), nil
}

func (s *Service) RegisterRunner(ctx context.Context, req *connect.ClientStream[ftlv1.RegisterRunnerRequest]) (*connect.Response[ftlv1.RegisterRunnerResponse], error) {
	logger := log.FromContext(ctx)
	// Initial endpoint creation.
	if !req.Receive() {
		if req.Err() != nil {
			return nil, errors.WithStack(req.Err())
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("initial registration not received"))
	}
	msg := req.Msg()
	endpoint, err := url.Parse(msg.Endpoint)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid endpoint"))
	}
	if endpoint.Scheme != "http" && endpoint.Scheme != "https" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Errorf("invalid endpoint scheme %q", endpoint.Scheme))
	}
	runnerKey, err := ulid.Parse(msg.Key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "invalid key"))
	}

	// Check if we can contact the runner.
	client := rpc.Dial(ftlv1connect.NewRunnerServiceClient, endpoint.String(), log.Error)
	retry := backoff.Backoff{}
	err = rpc.Wait(ctx, retry, client)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.Wrap(err, "failed to connect to runner"))
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
	}) //
	if errors.Is(err, dal.ErrConflict) {
		return nil, connect.NewError(connect.CodeAlreadyExists, err)
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() {
		err := s.dal.DeregisterRunner(context.Background(), runnerKey)
		if err != nil {
			logger.Errorf(err, "Failed to Deregister runner %s", endpoint)
		} else {
			logger.Infof("Deregistered runner %s", endpoint)
		}
	}()

	runnerStr := fmt.Sprintf("%s (%s)", endpoint, runnerKey)

	logger.Infof("New runner %s", runnerStr)

	// Start receiving heartbeats from runner.
	heartbeat := make(chan *ftlv1.RegisterRunnerRequest)
	ctx = concurrency.Call(ctx, func() error {
		for req.Receive() {
			select {
			case heartbeat <- req.Msg():
			case <-ctx.Done():
				return nil
			}
		}
		if req.Err() != nil {
			logger.Errorf(req.Err(), "Heartbeat error from runner %s", runnerStr)
			return errors.WithStack(req.Err())
		}
		return nil
	})

	// Loop until we receive a heartbeat or the context is cancelled.
	for {
		select {
		case msg := <-heartbeat:
			logger.Tracef("Heartbeat received from runner %s", runnerStr)
			maybeDeployment, err := msg.DeploymentAsOptional()
			if err != nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
			}
			fromState, err := s.dal.GetRunnerState(ctx, runnerKey)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get runner state")
			}
			toState := dal.RunnerStateFromProto(msg.State)

			// Runner requesting a transition from claimed to idle just means
			// that the Runner hasn't received the reservation yet.
			if fromState == dal.RunnerStateClaimed && toState == dal.RunnerStateIdle {
				toState = dal.RunnerStateClaimed
			}
			err = s.dal.UpsertRunner(ctx, dal.Runner{
				Key:        runnerKey,
				Language:   msg.Language,
				Endpoint:   msg.Endpoint,
				State:      toState,
				Deployment: maybeDeployment,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to set runner state")
			}

		case <-time.After(s.heartbeatTimeout):
			err := connect.NewError(connect.CodeDeadlineExceeded, errors.New("heartbeat timeout"))
			logger.Errorf(err, "Heartbeat timeout from runner %s", runnerStr)
			return nil, err

		case <-ctx.Done():
			err := context.Cause(ctx)
			if errors.Is(err, context.Canceled) {
				err = nil
			}
			if err != nil {
				logger.Errorf(err, "Context cancelled for runner %s heartbeat", runnerStr)
				return nil, connect.NewError(connect.CodeAborted, errors.Wrap(err, "heartbeat cancelled"))
			}
			return connect.NewResponse(&ftlv1.RegisterRunnerResponse{}), nil
		}
	}
}

func (s *Service) GetDeployment(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentRequest]) (*connect.Response[ftlv1.GetDeploymentResponse], error) {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetDeploymentResponse{
		Schema:    deployment.Schema.ToProto().(*pschema.Module), //nolint:forcetypeassert
		Artefacts: slices.Map(deployment.Artefacts, func(artefact *dal.Artefact) *ftlv1.DeploymentArtefact { return artefact.ToProto() }),
	}), nil
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	deployment, err := s.getDeployment(ctx, req.Msg.DeploymentKey)
	if err != nil {
		return err
	}
	chunk := make([]byte, s.artefactChunkSize)
nextArtefact:
	for _, artefact := range deployment.Artefacts {
		for _, clientArtefact := range req.Msg.HaveArtefacts {
			if proto.Equal(artefact.ToProto(), clientArtefact) {
				continue nextArtefact
			}
		}
		for {
			n, err := artefact.Content.Read(chunk)
			if n != 0 {
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: artefact.ToProto(),
					Chunk:    chunk[:n],
				}); err != nil {
					return errors.Wrap(err, "could not send artefact chunk")
				}
			}
			if err == io.EOF {
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
	endpoints, err := s.dal.GetRoutingTable(ctx, req.Msg.Verb.Module)
	if err != nil {
		if errors.Is(err, dal.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.Wrapf(err, "no runners for module %q", req.Msg.Verb.Module))
		}
		return nil, errors.Wrap(err, "failed to get runners for module")
	}
	endpoint := endpoints[rand.Intn(len(endpoints))] //nolint:gosec
	client := s.clientsForEndpoint(endpoint)
	resp, err := client.verb.Call(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}

func (s *Service) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	panic("unimplemented")
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
	module, err := schema.ModuleFromProto(ms)
	if err != nil {
		return nil, errors.Wrap(err, "invalid module schema")
	}
	key, err := s.dal.CreateDeployment(ctx, ms.Runtime.Language, module, artefacts)
	if err != nil {
		return nil, errors.Wrap(err, "could not create deployment")
	}
	logger.Infof("Created deployment %s", key)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: key.String()}), nil
}

func (s *Service) getDeployment(ctx context.Context, key string) (*dal.Deployment, error) {
	dkey, err := ulid.Parse(key)
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
		count, err := s.dal.DeleteStaleRunners(context.Background(), s.heartbeatTimeout)
		if err != nil {
			logger.Errorf(err, "Failed to delete stale runners")
		} else if count > 0 {
			logger.Warnf("Deleted %d stale runners", count)
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
		count, err := s.dal.ExpireRunnerReservations(ctx)
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
