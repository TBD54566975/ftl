package backplane

import (
	"context"
	"io"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backplane/internal/dao"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/server"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/common/socket"
	"github.com/TBD54566975/ftl/console"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
)

type Config struct {
	Bind socket.Socket `help:"Socket to bind to." default:"tcp://localhost:8892"`
	DSN  string        `help:"Postgres DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret"`
}

func Start(ctx context.Context, config Config) error {
	logger := log.FromContext(ctx)
	logger.Infof("Starting FTL backplane")
	conn, err := pgxpool.New(ctx, config.DSN)
	if err != nil {
		return nil
	}

	c, err := console.Server(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("Listening on %s", config.Bind.URL())

	svc := &Service{dao: dao.New(conn)}

	reflector := grpcreflect.NewStaticReflector(ftlv1connect.BackplaneServiceName)
	return server.Serve(ctx, config.Bind,
		server.GRPC(ftlv1connect.NewVerbServiceHandler, svc),
		server.GRPC(ftlv1connect.NewBackplaneServiceHandler, svc),
		server.Route(grpcreflect.NewHandlerV1(reflector)),
		server.Route(grpcreflect.NewHandlerV1Alpha(reflector)),
		server.Route("/", c),
	)
}

var _ ftlv1connect.BackplaneServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type Service struct {
	dao *dao.DAO
}

func (s *Service) GetDeploymentArtefacts(ctx context.Context, req *connect.Request[ftlv1.GetDeploymentArtefactsRequest], resp *connect.ServerStream[ftlv1.GetDeploymentArtefactsResponse]) error {
	dkey, err := uuid.Parse(req.Msg.DeploymentKey)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, errors.WithStack(err))
	}
	deployment, err := s.dao.GetDeployment(ctx, dkey)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, errors.WithStack(err))
	}
	haveDigests := mapset.NewSet(req.Msg.HaveDigests...)
	chunk := make([]byte, 1024*1024)
	for _, artefact := range deployment.Artefacts {
		if haveDigests.Contains(artefact.Digest.String()) {
			continue
		}
		for {
			n, err := artefact.Content.Read(chunk)
			if n != 0 {
				if err := resp.Send(&ftlv1.GetDeploymentArtefactsResponse{
					Artefact: artefact.ToProto(),
					Chunk:    chunk[:n],
				}); err != nil {
					return errors.WithStack(err)
				}
			}
			if err == io.EOF {
				break
			} else if err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	panic("unimplemented")
}

func (s *Service) List(ctx context.Context, req *connect.Request[ftlv1.ListRequest]) (*connect.Response[ftlv1.ListResponse], error) {
	panic("unimplemented")
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	byteDigests, err := slices.MapErr(req.Msg.ClientDigests, sha256.ParseSHA256)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	need, err := s.dao.GetMissingArtefacts(ctx, byteDigests)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: slices.Map(need, func(s sha256.SHA256) string { return s.String() }),
	}), nil
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	msg := req.Msg
	digest, err := s.dao.CreateArtefact(ctx, msg.Content)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: digest[:]}), nil
}

func (s *Service) CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	logger := log.FromContext(ctx)
	artefacts, err := slices.MapErr(req.Msg.Artefacts, func(in *ftlv1.DeploymentArtefact) (dao.DeploymentArtefact, error) {
		digest, err := sha256.ParseSHA256(in.Digest)
		if err != nil {
			return dao.DeploymentArtefact{}, errors.Wrap(err, "invalid digest")
		}
		return dao.DeploymentArtefact{
			Executable: in.Executable,
			Path:       in.Path,
			Digest:     digest,
		}, nil
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ms := req.Msg.Schema
	key, err := s.dao.CreateDeployment(ctx, ms.Runtime.Language, schema.ModuleFromProto(ms), artefacts)
	if err != nil {
		return nil, errors.Wrap(err, "could not create deployment")
	}
	logger.Infof("Created deployment %s", key)
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: key.String()}), nil
}
