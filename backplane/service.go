package backplane

import (
	"context"
	"io"

	"github.com/alecthomas/errors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backplane/internal/dao"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

var _ ftlv1.BackplaneServiceServer = (*Service)(nil)
var _ ftlv1.VerbServiceServer = (*Service)(nil)

func New(ctx context.Context, dsn string) (*Service, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Service{dao: dao.New(conn)}, nil
}

type Service struct {
	dao *dao.DAO
}

func (s *Service) Ping(ctx context.Context, req *ftlv1.PingRequest) (*ftlv1.PingResponse, error) {
	return &ftlv1.PingResponse{}, nil
}

func (s *Service) Call(ctx context.Context, req *ftlv1.CallRequest) (*ftlv1.CallResponse, error) {
	panic("unimplemented")
}

func (s *Service) List(ctx context.Context, req *ftlv1.ListRequest) (*ftlv1.ListResponse, error) {
	panic("unimplemented")
}

func (s *Service) Send(ctx context.Context, req *ftlv1.SendRequest) (*ftlv1.SendResponse, error) {
	panic("unimplemented")
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *ftlv1.GetArtefactDiffsRequest) (*ftlv1.GetArtefactDiffsResponse, error) {
	panic("unimplemented")
}

func (s *Service) UploadArtefact(stream ftlv1.BackplaneService_UploadArtefactServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return errors.WithStack(err)
		}
		_, err = s.dao.CreateArtefact(stream.Context(), req.Path, req.Executable, req.Content)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
