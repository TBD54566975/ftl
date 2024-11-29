package proxy

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftlv1connect.ModuleServiceHandler = &Service{}

type Service struct {
	controllerVerbService   ftlv1connect.VerbServiceClient
	controllerModuleService ftlv1connect.ModuleServiceClient
}

func New(controllerVerbService ftlv1connect.VerbServiceClient, controllerModuleService ftlv1connect.ModuleServiceClient) *Service {
	proxy := &Service{
		controllerVerbService:   controllerVerbService,
		controllerModuleService: controllerModuleService,
	}
	return proxy
}

func (r *Service) GetModuleContext(ctx context.Context, c *connect.Request[ftlv1.GetModuleContextRequest], c2 *connect.ServerStream[ftlv1.GetModuleContextResponse]) error {
	moduleContext, err := r.controllerModuleService.GetModuleContext(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return fmt.Errorf("failed to get module context: %w", err)
	}
	for {
		rcv := moduleContext.Receive()
		if rcv {
			err := c2.Send(moduleContext.Msg())
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
		} else if moduleContext.Err() != nil {
			return fmt.Errorf("failed to receive message: %w", moduleContext.Err())
		}
	}

}

func (r *Service) AcquireLease(ctx context.Context, c *connect.BidiStream[ftlv1.AcquireLeaseRequest, ftlv1.AcquireLeaseResponse]) error {
	lease := r.controllerModuleService.AcquireLease(ctx)
	for {
		req, err := c.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive message: %w", err)
		}
		err = lease.Send(req)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		msg, err := lease.Receive()
		if err != nil {
			return fmt.Errorf("failed to receive response message: %w", err)
		}
		err = c.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send response message: %w", err)
		}
	}

}

func (r *Service) PublishEvent(ctx context.Context, c *connect.Request[ftlv1.PublishEventRequest]) (*connect.Response[ftlv1.PublishEventResponse], error) {
	event, err := r.controllerModuleService.PublishEvent(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to proxy event: %w", err)
	}
	return event, nil
}

func (r *Service) Ping(ctx context.Context, c *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (r *Service) Call(ctx context.Context, c *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {

	call, err := r.controllerVerbService.Call(ctx, headers.CopyRequestForForwarding(c))
	if err != nil {
		return nil, fmt.Errorf("failed to proxy verb: %w", err)
	}
	return call, nil
}
