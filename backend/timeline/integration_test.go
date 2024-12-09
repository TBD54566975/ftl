//go:build integration

package timeline

import (
	"context"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestTimeline(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go"),
		in.CopyModule("cron"),
		in.CopyModule("time"),
		in.CopyModule("echo"),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.CopyModule("ingress"),
		in.Deploy("cron"),
		in.Deploy("time"),
		in.Deploy("echo"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),
		in.Deploy("ingress"),

		// Trigger events
		in.HttpCall(http.MethodGet, "/users/123/posts/456", nil, nil, func(t testing.TB, resp *in.HTTPResponse) {}),
		in.Call("echo", "echo", in.Obj{}, func(t testing.TB, response in.Obj) {}),
		in.Call("publisher", "publish", in.Obj{}, func(t testing.TB, resp in.Obj) {}),

		in.SubTests(
			in.SubTest{Name: "Limit", Action: in.VerifyTimeline(1, []*timelinepb.GetTimelineRequest_Filter{}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				assert.Equal(t, 1, len(events))
			})},
			in.SubTest{Name: "IngressEvent", Action: in.VerifyTimeline(1000, []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_INGRESS},
						},
					},
				},
			}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				assert.Equal(t, 1, len(events))
				ingress, ok := events[0].Entry.(*timelinepb.Event_Ingress)
				assert.True(t, ok, "expected ingress event")

				assert.Equal(t, ingress.Ingress.VerbRef.Module, "ingress")
				assert.Equal(t, ingress.Ingress.VerbRef.Name, "get")
			})},
			in.SubTest{Name: "CallEvent", Action: in.VerifyTimeline(1000, []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_CALL},
						},
					},
				},
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_Call{
						Call: &timelinepb.GetTimelineRequest_CallFilter{
							DestModule: "echo",
						},
					},
				},
			}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				assert.Equal(t, 1, len(events))
				call, ok := events[0].Entry.(*timelinepb.Event_Call)
				assert.True(t, ok, "expected call event")

				assert.Equal(t, call.Call.DestinationVerbRef.Module, "echo")
				assert.Equal(t, call.Call.DestinationVerbRef.Name, "echo")
				assert.Contains(t, call.Call.Response, "Hello, world!!!")
			})},
			in.SubTest{Name: "CronEvent", Action: in.VerifyTimeline(1, []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_CRON_SCHEDULED},
						},
					},
				},
			}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				scheduled, ok := events[0].Entry.(*timelinepb.Event_CronScheduled)
				assert.True(t, ok, "expected scheduled event")

				assert.Equal(t, scheduled.CronScheduled.VerbRef.Module, "cron")
				assert.Equal(t, scheduled.CronScheduled.VerbRef.Name, "job")
			})},
			in.SubTest{Name: "PublishEvent", Action: in.VerifyTimeline(1000, []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_PUBSUB_PUBLISH},
						},
					},
				},
			}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				assert.Equal(t, 1, len(events))
				publish, ok := events[0].Entry.(*timelinepb.Event_PubsubPublish)
				assert.True(t, ok, "expected publish event")

				assert.Equal(t, publish.PubsubPublish.Topic, "testTopic")
				assert.Equal(t, publish.PubsubPublish.VerbRef.Module, "publisher")
				assert.Equal(t, publish.PubsubPublish.VerbRef.Name, "publish")
			})},
			in.SubTest{Name: "ConsumeEvent", Action: in.VerifyTimeline(1000, []*timelinepb.GetTimelineRequest_Filter{
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
						EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
							EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME},
						},
					},
				},
			}, func(ctx context.Context, t testing.TB, events []*timelinepb.Event) {
				assert.Equal(t, 1, len(events))
				consume, ok := events[0].Entry.(*timelinepb.Event_PubsubConsume)
				assert.True(t, ok, "expected consume event")

				assert.Equal(t, *consume.PubsubConsume.DestVerbModule, "subscriber")
				assert.Equal(t, *consume.PubsubConsume.DestVerbName, "consume")
			})},
			in.SubTest{Name: "DeleteOldEvents", Action: in.DeleteOldTimelineEvents(1, timelinepb.EventType_EVENT_TYPE_INGRESS,
				func(ctx context.Context, t testing.TB, expectDeleted int, events []*timelinepb.Event) {
					assert.Equal(t, expectDeleted, 1)
					for _, event := range events {
						_, ok := event.Entry.(*timelinepb.Event_Ingress)
						if ok {
							t.Errorf("expected no ingress events, got %v", event.Entry)
						}
					}
				}),
			},
		),
	)
}

type streamState struct {
	ascEvents     []*timelinepb.Event
	descEvents    []*timelinepb.Event
	actualEntries []*timelinepb.CreateEventsRequest_EventEntry
}

func TestStreamTimeline(t *testing.T) {
	lock := &sync.Mutex{}
	state := &streamState{}
	in.Run(t,
		// stream events into two slices, one in ascending order and one in descending order
		streamEvents(60, timelinepb.GetTimelineRequest_ORDER_ASC, state, lock),
		streamEvents(60, timelinepb.GetTimelineRequest_ORDER_DESC, state, lock),

		// create events with timestamps out of order, simulating different services publishing events at different times
		createOutOfOrderEvents(100, state),
		in.Sleep(3*time.Second),
		createOutOfOrderEvents(100, state),
		in.Sleep(3*time.Second),

		// check that all events where streamed and in the correct order
		checkEvents(state, true, lock),
		checkEvents(state, false, lock),
	)
}

func streamEvents(pageSize int32, order timelinepb.GetTimelineRequest_Order, state *streamState, lock *sync.Mutex) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Steaming events (order = %v)", order)
		go func() {
			stream, err := ic.Timeline.StreamTimeline(ic.Context, connect.NewRequest(&timelinepb.StreamTimelineRequest{
				Query: &timelinepb.GetTimelineRequest{
					Limit: pageSize,
					Order: order,
				},
			}))
			assert.NoError(t, err)
			defer stream.Close()
			for stream.Receive() {
				lock.Lock()
				log.FromContext(ic.Context).Infof("streamed %d events", len(stream.Msg().Events))
				if order == timelinepb.GetTimelineRequest_ORDER_ASC {
					state.ascEvents = append(state.ascEvents, stream.Msg().Events...)
				} else {
					reverseEvents := make([]*timelinepb.Event, 0, len(stream.Msg().Events))
					reverseEvents = append(reverseEvents, stream.Msg().Events...)
					slices.Reverse(reverseEvents)
					state.descEvents = append(state.descEvents, reverseEvents...)
				}
				lock.Unlock()
			}
		}()
	}
}

func createOutOfOrderEvents(count int, state *streamState) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Creating events")
		for i := range count {
			entry := &timelinepb.CreateEventsRequest_EventEntry{
				Timestamp: timestamppb.New(time.Now()),
				Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated{
					DeploymentCreated: &timelinepb.DeploymentCreatedEvent{
						Key:         "fake",
						Language:    "go",
						ModuleName:  "fakemodule:" + strconv.Itoa(i),
						MinReplicas: 1,
					},
				},
			}
			_, err := ic.Timeline.CreateEvents(ic.Context, connect.NewRequest(&timelinepb.CreateEventsRequest{
				Entries: []*timelinepb.CreateEventsRequest_EventEntry{
					entry,
				},
			}))
			assert.NoError(t, err)
			state.actualEntries = append(state.actualEntries, entry)
		}
	}
}

func checkEvents(state *streamState, asc bool, lock *sync.Mutex) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("Checking events (asc = %v)", asc)
		lock.Lock()
		defer lock.Unlock()

		var streamEvents []*timelinepb.Event
		if asc {
			streamEvents = state.ascEvents
		} else {
			streamEvents = state.descEvents
		}
		assert.Equal(t, len(state.actualEntries), len(streamEvents), "expected all events to have been streamed")
		for i, event := range streamEvents {
			expectedEntry := state.actualEntries[i]
			assert.Equal(t, expectedEntry.GetDeploymentCreated().ModuleName, event.GetDeploymentCreated().ModuleName, "expected streamed event to match publication order")
		}
	}
}
