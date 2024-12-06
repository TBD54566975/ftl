package timeline

import (
	"context"

	"connectrpc.com/connect"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

//go:sumtype
type Event interface {
	ToReq() (*timelinepb.CreateEventRequest, error)
	clientEvent()
}

func Publish(ctx context.Context, event Event) {
	client := rpc.ClientFromContext[timelinev1connect.TimelineServiceClient](ctx)
	req, err := event.ToReq()
	if err != nil {
		log.FromContext(ctx).Warnf("failed to create request to publish %T event: %v", event, err)
		return
	}
	_, err = client.CreateEvent(ctx, connect.NewRequest(req))
	if err != nil {
		log.FromContext(ctx).Warnf("failed to publish %T event: %v", event, err)
	}
}
