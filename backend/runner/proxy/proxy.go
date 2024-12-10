package proxy

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/result"

	"github.com/TBD54566975/ftl/backend/controller/observability"
	ftldeployment "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1"
	ftldeploymentconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/deployment/v1/ftlv1connect"
	ftllease "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1"
	ftlleaseconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	ftlpubsubv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1"
	ftlv1connect2 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/pubsub/v1/ftlv1connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/rpc/headers"
	"github.com/TBD54566975/ftl/internal/schema"
)

var _ ftlv1connect.VerbServiceHandler = &Service{}
var _ ftldeploymentconnect.DeploymentServiceHandler = &Service{}

type moduleVerbService struct {
	client     ftlv1connect.VerbServiceClient
	deployment model.DeploymentKey
}

type Service struct {
	controllerDeploymentService ftldeploymentconnect.DeploymentServiceClient
	controllerLeaseService      ftlleaseconnect.LeaseServiceClient
	controllerPubsubService     ftlv1connect2.LegacyPubsubServiceClient
	moduleVerbService           map[string]moduleVerbService
}

func New(controllerModuleService ftldeploymentconnect.DeploymentServiceClient, leaseClient ftlleaseconnect.LeaseServiceClient, controllerPubsubService ftlv1connect2.LegacyPubsubServiceClient) *Service {
	proxy := &Service{
		controllerDeploymentService: controllerModuleService,
		controllerLeaseService:      leaseClient,
		controllerPubsubService:     controllerPubsubService,
		moduleVerbService:           map[string]moduleVerbService{},
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
			newRouteTable := map[string]moduleVerbService{}
			for _, route := range moduleContext.Msg().Routes {
				logger.Debugf("Adding route: %s -> %s", route.Deployment, route.Uri)

				deploment, err := model.ParseDeploymentKey(route.Deployment)
				if err != nil {
					return fmt.Errorf("failed to parse deployment key: %w", err)
				}
				module := deploment.Payload.Module
				if existing, ok := r.moduleVerbService[module]; !ok || existing.deployment.String() != deploment.String() {
					newRouteTable[module] = moduleVerbService{
						client:     rpc.Dial(ftlv1connect.NewVerbServiceClient, route.Uri, log.Error),
						deployment: deploment,
					}
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
	_, err := r.controllerLeaseService.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err != nil {
		return fmt.Errorf("failed to ping lease service: %w", err)
	}
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
			return connect.NewError(connect.CodeOf(err), fmt.Errorf("lease failed %w", err))
		}
		err = c.Send(msg)
		if err != nil {
			return fmt.Errorf("failed to send response message: %w", err)
		}
	}

}

func (r *Service) PublishEvent(ctx context.Context, c *connect.Request[ftlpubsubv1.PublishEventRequest]) (*connect.Response[ftlpubsubv1.PublishEventResponse], error) {
	event, err := r.controllerPubsubService.PublishEvent(ctx, connect.NewRequest(c.Msg))
	if err != nil {
		return nil, fmt.Errorf("failed to proxy event: %w", err)
	}
	return event, nil
}

func (r *Service) Ping(ctx context.Context, c *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (r *Service) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	start := time.Now()
	verbService, ok := r.moduleVerbService[req.Msg.Verb.Module]
	if !ok {
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("failed to find deployment for module"))
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("deployment not found"))
	}
	timelineClient := timeline.ClientFromContext(ctx)

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
		DeploymentKey: verbService.deployment,
		RequestKey:    requestKey,
		StartTime:     start,
		DestVerb:      schema.RefFromProto(req.Msg.Verb),
		Callers:       callers,
		Request:       req.Msg,
	}

	originalResp, err := verbService.client.Call(ctx, headers.CopyRequestForForwarding(req))
	if err != nil {
		callEvent.Response = result.Err[*ftlv1.CallResponse](err)
		timelineClient.Publish(ctx, callEvent)
		observability.Calls.Request(ctx, req.Msg.Verb, start, optional.Some("verb call failed"))
		return nil, fmt.Errorf("failed to proxy verb: %w", err)
	}
	resp := connect.NewResponse(originalResp.Msg)
	callEvent.Response = result.Ok(resp.Msg)
	timelineClient.Publish(ctx, callEvent)
	observability.Calls.Request(ctx, req.Msg.Verb, start, optional.None[string]())
	return resp, nil
}

// ResetSubscription is legacy, it will go once the DB based pubsub is removed
func (r *Service) ResetSubscription(ctx context.Context, req *connect.Request[ftlpubsubv1.ResetSubscriptionRequest]) (*connect.Response[ftlpubsubv1.ResetSubscriptionResponse], error) {
	return connect.NewResponse(&ftlpubsubv1.ResetSubscriptionResponse{}), nil
}
