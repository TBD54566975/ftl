package routing

import (
	"context"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

// RouteClientManager managed clients for the routing service, so calls to a given module can be routed to the correct instance.
type RouteClientManager[T rpc.Pingable] struct {
	routingTable  *RouteTable
	moduleClients *xsync.MapOf[string, optional.Option[T]]
	factory       rpc.ClientFactory[T]
}

func NewClientManager[T rpc.Pingable](ctx context.Context, routingTable *RouteTable, factory rpc.ClientFactory[T]) *RouteClientManager[T] {
	svc := &RouteClientManager[T]{
		routingTable:  routingTable,
		moduleClients: xsync.NewMapOf[string, optional.Option[T]](),
		factory:       factory,
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

func (s *RouteClientManager[T]) LookupClient(module string) optional.Option[T] {
	res, _ := s.moduleClients.LoadOrCompute(module, func() optional.Option[T] {
		route, ok := s.routingTable.Current().GetForModule(module).Get()
		if !ok {
			return optional.None[T]()
		}
		return optional.Some[T](rpc.Dial(s.factory, route.String(), log.Error))
	})
	return res
}
