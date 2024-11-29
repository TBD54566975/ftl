package pubsub

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	pb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1"
	pbconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/publish/v1/publishpbconnect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/schema"
	sl "github.com/TBD54566975/ftl/internal/slices"
)

type Service struct {
	publishers map[string]*publisher
}

var _ pbconnect.PublishServiceHandler = (*Service)(nil)

func New(module *schema.Module) (*Service, error) {
	publishers := map[string]*publisher{}
	for t := range sl.FilterVariants[*schema.Topic](module.Decls) {
		publisher, err := newPublisher(t)
		if err != nil {
			return nil, err
		}
		publishers[t.Name] = publisher
	}

	// TODO: topic producers needs to be closed eventually

	svc := &Service{
		publishers: publishers,
	}
	return svc, nil
}

func (s *Service) PublishEvent(ctx context.Context, req *connect.Request[pb.PublishEventRequest]) (*connect.Response[pb.PublishEventResponse], error) {
	publisher, ok := s.publishers[req.Msg.Topic.Name]
	if !ok {
		return nil, fmt.Errorf("topic %s not found", req.Msg.Topic.Name)
	}
	err := publisher.publish(req.Msg.Body, req.Msg.Key)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.PublishEventResponse{}), nil
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}
