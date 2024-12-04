package proxy

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftldeployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	ftldeploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	ftllease "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1"
	ftlleaseconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftldeploymentconnect.DeploymentServiceHandler = &Service{}

type Service struct {
	controllerVerbService       ftlv1connect.VerbServiceClient
	controllerDeploymentService ftldeploymentconnect.DeploymentServiceClient
	controllerLeaseService      ftlleaseconnect.LeaseServiceClient
}

func New(controllerVerbService ftlv1connect.VerbServiceClient, controllerModuleService ftldeploymentconnect.DeploymentServiceClient, controllerLeaseClient ftlleaseconnect.LeaseServiceClient) *Service {
	proxy := &Service{
		controllerVerbService:       controllerVerbService,
		controllerDeploymentService: controllerModuleService,
		controllerLeaseService:      controllerLeaseClient,
	}
	return proxy
}

func (r *Service) GetDeploymentContext(ctx context.Context, c *connect.Request[ftldeployment.GetDeploymentContextRequest], c2 *connect.ServerStream[ftldeployment.GetDeploymentContextResponse]) error {
	moduleContext, err := r.controllerDeploymentService.GetDeploymentContext(ctx, connect.NewRequest(c.Msg))
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

func (r *Service) AcquireLease(ctx context.Context, c *connect.BidiStream[ftllease.AcquireLeaseRequest, ftllease.AcquireLeaseResponse]) error {
	lease := r.controllerLeaseService.AcquireLease(ctx)
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

func (r *Service) PublishEvent(ctx context.Context, c *connect.Request[ftldeployment.PublishEventRequest]) (*connect.Response[ftldeployment.PublishEventResponse], error) {
	event, err := r.controllerDeploymentService.PublishEvent(ctx, connect.NewRequest(c.Msg))
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
