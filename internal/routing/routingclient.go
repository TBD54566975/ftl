package routing

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

var _ CallClient = (*RouteClientManager)(nil)

type CallClient interface {
	Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

// RouteClientManager managed clients for the routing service, so calls to a given module can be routed to the correct instance.
type RouteClientManager struct {
	routingTable  *RouteTable
	moduleClients *xsync.MapOf[string, optional.Option[ftlv1connect.VerbServiceClient]]
}

func (s *RouteClientManager) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	client, ok := s.LookupClient(req.Msg.Verb.Module).Get()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("module not found"))
	}
	call, err := client.Call(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to call module %s: %w", req.Msg.Verb.Module, err)
	}
	return call, err
}

func NewClientManager(ctx context.Context, changes schemaeventsource.EventSource) *RouteClientManager {
	svc := &RouteClientManager{
		routingTable:  New(ctx, changes),
		moduleClients: xsync.NewMapOf[string, optional.Option[ftlv1connect.VerbServiceClient]](),
	}
	routeUpdates := svc.routingTable.Subscribe()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case module := <-routeUpdates:
				svc.moduleClients.Delete(module)
			}
		}
	}()
	return svc
}

func (s *RouteClientManager) LookupClient(module string) optional.Option[ftlv1connect.VerbServiceClient] {
	res, _ := s.moduleClients.LoadOrCompute(module, func() optional.Option[ftlv1connect.VerbServiceClient] {
		route, ok := s.routingTable.Current().GetForModule(module).Get()
		if !ok {
			return optional.None[ftlv1connect.VerbServiceClient]()
		}
		return optional.Some[ftlv1connect.VerbServiceClient](rpc.Dial(ftlv1connect.NewVerbServiceClient, route.String(), log.Error))
	})
	return res
}
