package timeline

import (
	"context"
	"net/url"

	"connectrpc.com/connect"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type timelineContextKey struct{}

type Client struct {
	timelinev1connect.TimelineServiceClient
}

func NewClient(endpoint *url.URL) *Client {
	c := rpc.Dial(timelinev1connect.NewTimelineServiceClient, endpoint.String(), log.Error)
	client := &Client{TimelineServiceClient: c}
	return client
}

func ContextWithClient(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, timelineContextKey{}, client)
}

func ClientFromContext(ctx context.Context) *Client {
	c, ok := ctx.Value(timelineContextKey{}).(*Client)
	if !ok {
		panic("Timeline client not found in context")
	}
	return c
}

//go:sumtype
type Event interface {
	ToReq() (*timelinepb.CreateEventRequest, error)
	clientEvent()
}

func (c *Client) Publish(ctx context.Context, event Event) {
	req, err := event.ToReq()
	if err != nil {
		log.FromContext(ctx).Warnf("failed to create request to publish %T event: %v", event, err)
		return
	}
	_, err = c.CreateEvent(ctx, connect.NewRequest(req))
	if err != nil {
		log.FromContext(ctx).Warnf("failed to publish %T event: %v", event, err)
	}
}
