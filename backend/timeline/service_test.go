package timeline

import (
	"context"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/alecthomas/assert/v2"
)

func TestGetTimelineWithLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service := &service{}

	// Create a bunch of entries
	entryCount := 100
	requests := []*timelinepb.CreateEventRequest{}
	for i := range entryCount {
		requests = append(requests, &timelinepb.CreateEventRequest{
			Entry: &timelinepb.CreateEventRequest_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		})
	}
	for _, request := range requests {
		_, err := service.CreateEvent(ctx, connect.NewRequest(request))
		assert.NoError(t, err)
	}

	// Test with different limits
	for _, limit := range []int32{
		0,
		10,
		33,
		110,
	} {
		resp, err := service.GetTimeline(ctx, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Order: timelinepb.GetTimelineRequest_ORDER_DESC,
			Limit: limit,
			Filters: []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{
								timelinepb.EventType_EVENT_TYPE_CALL,
							},
						},
					},
				},
			},
		}))
		if limit == 0 {
			assert.Error(t, err, "invalid_argument: limit must be > 0")
			continue
		}
		assert.NoError(t, err)
		if limit == 0 || limit > int32(entryCount) {
			assert.Equal(t, entryCount, len(resp.Msg.Events))
		} else {
			assert.Equal(t, int(limit), len(resp.Msg.Events))
		}
	}
}

func TestDeleteOldEvents(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service := &service{}

	// Create a bunch of entries of different types
	requests := []*timelinepb.CreateEventRequest{}
	for i := range 100 {
		requests = append(requests, &timelinepb.CreateEventRequest{
			Entry: &timelinepb.CreateEventRequest_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventRequest{
			Entry: &timelinepb.CreateEventRequest_Log{
				Log: &timelinepb.LogEvent{
					Message: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventRequest{
			Entry: &timelinepb.CreateEventRequest_DeploymentCreated{
				DeploymentCreated: &timelinepb.DeploymentCreatedEvent{
					Key: strconv.Itoa(i),
				},
			},
		})
	}
	for i, request := range requests {
		if i == 150 {
			// Add a delay half way through
			time.Sleep(3 * time.Second)
		}
		_, err := service.CreateEvent(ctx, connect.NewRequest(request))
		assert.NoError(t, err)
	}

	// Delete half the events (everything older than 3 seconds)
	resp, err := service.DeleteOldEvents(ctx, connect.NewRequest(&timelinepb.DeleteOldEventsRequest{
		AgeSeconds: 3,
		EventType:  timelinepb.EventType_EVENT_TYPE_UNSPECIFIED,
	}))
	assert.NoError(t, err)
	assert.Equal(t, len(service.events), 150, "expected only half the events to be deleted")
	assert.Equal(t, resp.Msg.DeletedCount, 150, "expected half the events to be in the deletion count")
}
