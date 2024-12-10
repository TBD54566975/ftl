package timeline

import (
	"context"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinev1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

const (
	maxBatchSize  = 16
	maxBatchDelay = 100 * time.Millisecond
)

type timelineContextKey struct{}

type Client struct {
	timelinev1connect.TimelineServiceClient

	entries          chan *timelinepb.CreateEventsRequest_EventEntry
	lastDroppedError atomic.Value[time.Time]
	lastFailedError  atomic.Value[time.Time]
}

func NewClient(ctx context.Context, endpoint *url.URL) *Client {
	c := rpc.Dial(timelinev1connect.NewTimelineServiceClient, endpoint.String(), log.Error)
	client := &Client{
		TimelineServiceClient: c,
		entries:               make(chan *timelinepb.CreateEventsRequest_EventEntry, 1000),
	}
	go client.processEvents(ctx)
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
	ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error)
	clientEvent()
}

// Publish asynchronously enqueues an event for publication to the timeline.
func (c *Client) Publish(ctx context.Context, event Event) {
	entry, err := event.ToEntry()
	entry.Timestamp = timestamppb.New(time.Now())
	if err != nil {
		log.FromContext(ctx).Warnf("failed to create request to publish %T event: %v", event, err)
		return
	}
	select {
	case c.entries <- entry:
	default:
		if time.Since(c.lastDroppedError.Load()) > 10*time.Second {
			log.FromContext(ctx).Warnf("Dropping event %T due to full queue", event)
			c.lastDroppedError.Store(time.Now())
		}
	}
}

func (c *Client) processEvents(ctx context.Context) {
	lastFlush := time.Now()
	buffer := make([]*timelinepb.CreateEventsRequest_EventEntry, 0, maxBatchSize)
	for {
		select {
		case <-ctx.Done():
			return

		case entry := <-c.entries:
			buffer = append(buffer, entry)

			if len(buffer) < maxBatchSize || time.Since(lastFlush) < maxBatchDelay {
				continue
			}
			c.flushEvents(ctx, buffer)
			buffer = nil

		case <-time.After(maxBatchDelay):
			if len(buffer) == 0 {
				continue
			}
			c.flushEvents(ctx, buffer)
			buffer = nil
		}
	}
}

// Flush all events in the buffer to the timeline service in a single call.
func (c *Client) flushEvents(ctx context.Context, entries []*timelinepb.CreateEventsRequest_EventEntry) {
	logger := log.FromContext(ctx).Scope("timeline")
	_, err := c.CreateEvents(ctx, connect.NewRequest(&timelinepb.CreateEventsRequest{
		Entries: entries,
	}))
	if err != nil {
		if time.Since(c.lastFailedError.Load()) > 10*time.Second {
			logger.Errorf(err, "Failed to insert %d events", len(entries))
			c.lastFailedError.Store(time.Now())
		}
		metrics.Failed(ctx, len(entries))
		return
	}
	metrics.Inserted(ctx, len(entries))
}
