package controlplane

import (
	"context"

	"github.com/alecthomas/errors"
	connect "github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
)

type ConsoleService struct {
	dal dal.DAL
}

var _ pbconsoleconnect.ConsoleServiceHandler = (*ConsoleService)(nil)

func NewConsoleService(dal dal.DAL) *ConsoleService {
	return &ConsoleService{
		dal: dal,
	}
}

func (*ConsoleService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (c *ConsoleService) GetModules(ctx context.Context, req *connect.Request[pbconsole.GetModulesRequest]) (*connect.Response[pbconsole.GetModulesResponse], error) {
	status, err := c.dal.GetStatus(ctx, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var modules []*pbconsole.Module
	for _, deployment := range status.Deployments {
		modules = append(modules, &pbconsole.Module{
			Name:     deployment.Module,
			Language: deployment.Language,
			Verbs:    []*pbconsole.Verb{},
		})
	}

	return connect.NewResponse(&pbconsole.GetModulesResponse{
		Modules: modules,
	}), nil
}
