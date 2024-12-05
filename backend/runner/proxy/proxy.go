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
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftldeploymentconnect.DeploymentServiceHandler = &Service{}

type Service struct {
	controllerDeploymentService ftldeploymentconnect.DeploymentServiceClient
	controllerLeaseService      ftlleaseconnect.LeaseServiceClient
	moduleVerbService           map[string]ftlv1connect.VerbServiceClient
}

func New(controllerModuleService ftldeploymentconnect.DeploymentServiceClient, leaseClient ftlleaseconnect.LeaseServiceClient) *Service {
	proxy := &Service{
		controllerDeploymentService: controllerModuleService,
		controllerLeaseService:      leaseClient,
		moduleVerbService:           map[string]ftlv1connect.VerbServiceClient{},
	}
	return proxy
}

func (r *Service) GetDeploymentContext(ctx context.Context, c *connect.Request[ftldeployment.GetDeploymentContextRequest], c2 *connect.ServerStream[ftldeployment.GetDeploymentContextResponse]) error {
	moduleContext, err := r.controllerDeploymentService.GetDeploymentContext(ctx, connect.NewRequest(c.Msg))
	logger := log.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module context: %w", err)
	}
	for {
		rcv := moduleContext.Receive()

		if rcv {
			logger.Debugf("Received DeploymentContext from module: %v", moduleContext.Msg())
			newRouteTable := map[string]ftlv1connect.VerbServiceClient{}
			for _, route := range moduleContext.Msg().Routes {
				logger.Debugf("Adding route: %s -> %s", route.Module, route.Uri)
				if client, ok := r.moduleVerbService[route.Module]; ok {
					newRouteTable[route.Module] = client
				} else {
					newRouteTable[route.Module] = rpc.Dial(ftlv1connect.NewVerbServiceClient, route.Uri, log.Error)
				}
			}
			r.moduleVerbService = newRouteTable
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
	defer lease.CloseResponse() //nolint:errcheck
	defer lease.CloseRequest()  //nolint:errcheck
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

	client, ok := r.moduleVerbService[c.Msg.Verb.Module]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("module not found in runners route table: %s", c.Msg.Verb.Module))
	}

	call, err := client.Call(ctx, headers.CopyRequestForForwarding(c))
	if err != nil {
		return nil, fmt.Errorf("failed to proxy verb: %w", err)
	}
	return call, nil
}
