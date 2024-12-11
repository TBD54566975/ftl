package routing

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/backend/controller/observability"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

var _ CallClient = (*VerbCallRouter)(nil)

type CallClient interface {
	Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

// VerbCallRouter managed clients for the routing service, so calls to a given module can be routed to the correct instance.
type VerbCallRouter struct {
	routingTable   *RouteTable
	moduleClients  *xsync.MapOf[string, optional.Option[ftlv1connect.VerbServiceClient]]
	timelineClient *timeline.Client
}

func (s *VerbCallRouter) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	start := time.Now()

	client, deployment, ok := s.LookupClient(req.Msg.Verb.Module)
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment for module"))
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("deployment not found"))
	}

	callers, err := headers.GetCallers(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get callers"))
		return nil, fmt.Errorf("could not get callers from headers: %w", err)
	}

	requestKey, ok, err := headers.GetRequestKey(req.Header())
	if err != nil {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to get request key"))
		return nil, fmt.Errorf("could not process headers for request key: %w", err)
	} else if !ok {
		requestKey = model.NewRequestKey(model.OriginIngress, "grpc")
		headers.SetRequestKey(req.Header(), requestKey)
	}

	callEvent := &timeline.Call{
		DeploymentKey: deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      schema.RefFromProto(req.Msg.Verb),
		Callers:       callers,
		Request:       req.Msg,
	}

	originalResp, err := client.Call(ctx, req)
	if err != nil {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		s.timelineClient.Publish(ctx, callEvent)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		return nil, fmt.Errorf("failed to call %s: %w", callEvent.DestVerb, err)
	}
	resp := connect.NewResponse(originalResp.Msg)
	callEvent.Response = result.Ok(resp.Msg)
	s.timelineClient.Publish(ctx, callEvent)
	observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	return resp, nil
}

func NewVerbRouterFromTable(ctx context.Context, routeTable *RouteTable, timelineClient *timeline.Client) *VerbCallRouter {
	svc := &VerbCallRouter{
		routingTable:   routeTable,
		moduleClients:  xsync.NewMapOf[string, optional.Option[ftlv1connect.VerbServiceClient]](),
		timelineClient: timelineClient,
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
func NewVerbRouter(ctx context.Context, changes schemaeventsource.EventSource, timelineClient *timeline.Client) *VerbCallRouter {
	return NewVerbRouterFromTable(ctx, New(ctx, changes), timelineClient)
}

func (s *VerbCallRouter) LookupClient(module string) (client ftlv1connect.VerbServiceClient, deployment model.DeploymentKey, ok bool) {
	res, _ := s.moduleClients.LoadOrCompute(module, func() optional.Option[ftlv1connect.VerbServiceClient] {
		current := s.routingTable.Current()
		var ok bool
		deployment, ok = current.GetDeployment(module).Get()
		if !ok {
			return optional.None[ftlv1connect.VerbServiceClient]()
		}
		route, ok := current.Get(deployment).Get()
		if !ok {
			return optional.None[ftlv1connect.VerbServiceClient]()
		}
		return optional.Some[ftlv1connect.VerbServiceClient](rpc.Dial(ftlv1connect.NewVerbServiceClient, route.String(), log.Error))
	})
	client, ok = res.Get()
	if !ok {
		return nil, deployment, false
	}
	return client, deployment, true
}
