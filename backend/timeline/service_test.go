package timeline

import (
	"context"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/alecthomas/assert/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetTimelineWithLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service := &service{}

	// Create a bunch of entries
	entryCount := 100
	entries := []*timelinepb.CreateEventsRequest_EventEntry{}
	for i := range entryCount {
		entries = append(entries, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamppb.New(time.Now()),
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		})
	}

	_, err := service.CreateEvents(ctx, connect.NewRequest(&timelinepb.CreateEventsRequest{
		Entries: entries,
	}))
	assert.NoError(t, err)

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
	entries := []*timelinepb.CreateEventsRequest_EventEntry{}
	for i := range 100 {
		var timestamp *timestamppb.Timestamp
		if i < 50 {
			timestamp = timestamppb.New(time.Now().Add(-3 * time.Second))
		} else {
			timestamp = timestamppb.New(time.Now())
		}
		entries = append(entries, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Call{
				Call: &timelinepb.CallEvent{
					Request:  strconv.Itoa(i),
					Response: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_Log{
				Log: &timelinepb.LogEvent{
					Message: strconv.Itoa(i),
				},
			},
		}, &timelinepb.CreateEventsRequest_EventEntry{
			Timestamp: timestamp,
			Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated{
				DeploymentCreated: &timelinepb.DeploymentCreatedEvent{
					Key: strconv.Itoa(i),
				},
			},
		})
	}

	_, err := service.CreateEvents(ctx, connect.NewRequest(&timelinepb.CreateEventsRequest{
		Entries: entries,
	}))
	assert.NoError(t, err)

	// Delete half the events (everything older than 3 seconds)
	_, err = service.DeleteOldEvents(ctx, connect.NewRequest(&timelinepb.DeleteOldEventsRequest{
		AgeSeconds: 3,
		EventType:  timelinepb.EventType_EVENT_TYPE_UNSPECIFIED,
	}))
	assert.NoError(t, err)
	assert.Equal(t, len(service.events), 150, "expected only half the events to be deleted")
}
