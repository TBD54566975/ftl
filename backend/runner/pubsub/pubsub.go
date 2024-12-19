package pubsub

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	pb "github.com/block/ftl/backend/protos/xyz/block/ftl/publish/v1"
	pbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/publish/v1/publishpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/schema"
	sl "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/model"
)

type Service struct {
	moduleName string
	publishers map[string]*publisher
	consumers  map[string]*consumer
}

type VerbClient interface {
	Call(ctx context.Context, c *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error)
}

var _ pbconnect.PublishServiceHandler = (*Service)(nil)

func New(module *schema.Module, deployment model.DeploymentKey, verbClient VerbClient, timelineClient *timeline.Client) (*Service, error) {
	publishers := map[string]*publisher{}
	for t := range sl.FilterVariants[*schema.Topic](module.Decls) {
		publisher, err := newPublisher(module.Name, t, deployment, timelineClient)
		if err != nil {
			return nil, err
		}
		publishers[t.Name] = publisher
	}

	consumers := map[string]*consumer{}
	for v := range sl.FilterVariants[*schema.Verb](module.Decls) {
		subscriber, ok := sl.FindVariant[*schema.MetadataSubscriber](v.Metadata)
		if !ok {
			continue
		}
		var deadLetterPublisher optional.Option[*publisher]
		if subscriber.DeadLetter {
			p, ok := publishers[schema.DeadLetterNameForSubscriber(v.Name)]
			if !ok {
				return nil, fmt.Errorf("dead letter publisher not found for subscription %s", v.Name)
			}
			deadLetterPublisher = optional.Some(p)
		}
		consumer, err := newConsumer(module.Name, v, subscriber, deployment, deadLetterPublisher, verbClient, timelineClient)
		if err != nil {
			return nil, err
		}
		consumers[v.Name] = consumer
	}

	return &Service{
		moduleName: module.Name,
		publishers: publishers,
		consumers:  consumers,
	}, nil
}

func (s *Service) Consume(ctx context.Context) error {
	for _, c := range s.consumers {
		err := c.Begin(ctx)
		if err != nil {
			return fmt.Errorf("could not begin consumer: %w", err)
		}
	}
	return nil
}

func (s *Service) PublishEvent(ctx context.Context, req *connect.Request[pb.PublishEventRequest]) (*connect.Response[pb.PublishEventResponse], error) {
	publisher, ok := s.publishers[req.Msg.Topic.Name]
	if !ok {
		return nil, fmt.Errorf("topic %s not found", req.Msg.Topic.Name)
	}
	caller := schema.Ref{
		Module: s.moduleName,
		Name:   req.Msg.Caller,
	}
	err := publisher.publish(ctx, req.Msg.Body, req.Msg.Key, caller)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.PublishEventResponse{}), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
