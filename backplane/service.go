package backplane

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backplane/internal/dao"
	v1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var _ ftlv1connect.BackplaneServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

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

func (s *Service) Call(ctx context.Context, req *connect.Request[v1.CallRequest]) (*connect.Response[v1.CallResponse], error) {
	panic("unimplemented")
}

func (s *Service) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	panic("unimplemented")
}

func (s *Service) Send(ctx context.Context, req *connect.Request[v1.SendRequest]) (*connect.Response[v1.SendResponse], error) {
	panic("unimplemented")
}

func (s *Service) GetArtefactDiffs(ctx context.Context, req *connect.Request[v1.GetArtefactDiffsRequest]) (*connect.Response[v1.GetArtefactDiffsResponse], error) {
	panic("unimplemented")
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	panic("unimplemented")
}

func (s *Service) UploadArtefact(ctx context.Context, req *connect.ClientStream[v1.UploadArtefactRequest]) (*connect.Response[v1.UploadArtefactResponse], error) {
	panic("unimplemented")
}

// func (s *Service) UploadArtefact(stream ftlv1.BackplaneService_UploadArtefactServer) error {
// 	for {
// 		req, err := stream.Recv()
// 		if errors.Is(err, io.EOF) {
// 			break
// 		} else if err != nil {
// 			return errors.WithStack(err)
// 		}
// 		_, err = s.dao.CreateArtefact(stream.Context(), req.Path, req.Executable, req.Content)
// 		if err != nil {
// 			return errors.WithStack(err)
// 		}
// 	}
// 	return nil
// }
