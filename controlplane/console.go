package controlplane

import (
	"context"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	connect "github.com/bufbuild/connect-go"
)

type ConsoleService struct {
	dal dal.DAL
}

var _ ftlv1connect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewConsoleService(dal dal.DAL) *ConsoleService {
	return &ConsoleService{
		dal: dal,
	}
}

func (*ConsoleService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (c *ConsoleService) GetModules(context.Context, *connect.Request[ftlv1.GetModulesRequest]) (*connect.Response[ftlv1.GetModulesResponse], error) {

	return connect.NewResponse(&ftlv1.GetModulesResponse{}), nil
}
