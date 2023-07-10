package controlplane

import (
	"context"

	"github.com/alecthomas/errors"
	connect "github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/controlplane/internal/dal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
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

func (c *ConsoleService) GetModules(ctx context.Context, req *connect.Request[ftlv1.GetModulesRequest]) (*connect.Response[ftlv1.GetModulesResponse], error) {
	status, err := c.dal.GetStatus(ctx, true)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var modules []*ftlv1.Module
	for _, deployment := range status.Deployments {
		modules = append(modules, &ftlv1.Module{
			Name:     deployment.Module,
			Language: deployment.Language,
			Verbs:    []*ftlv1.Verb{},
		})
	}

	return connect.NewResponse(&ftlv1.GetModulesResponse{
		Modules: modules,
	}), nil
}
