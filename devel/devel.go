package devel

import (
	"context"

	"github.com/bufbuild/connect-go"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var _ ftlv1connect.RunnerServiceHandler = (*Service)(nil)
var _ ftlv1connect.VerbServiceHandler = (*Service)(nil)

type Service struct {
}

func (s *Service) Call(ctx context.Context, c *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	panic("implement me")
}

func (s *Service) Ping(ctx context.Context, c *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	panic("implement me")
}

func (s *Service) DeployToRunner(ctx context.Context, c *connect.Request[ftlv1.DeployToRunnerRequest]) (*connect.Response[ftlv1.DeployToRunnerResponse], error) {
	panic("implement me")
}
