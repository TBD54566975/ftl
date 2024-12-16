//go:build integration

package pubsub

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/common/slices"
	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/model"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestPubSub(t *testing.T) {
	calls := 20
	events := calls * 10
	in.Run(t,
		in.WithLanguages("java", "go", "kotlin"),
		in.WithPubSub(),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Repeat(calls, in.Call("publisher", "publishTen", in.Obj{}, func(t testing.TB, resp in.Obj) {})),

		in.Sleep(time.Second*4),

		// check that there are the right amount of successful async calls
		checkConsumed("subscriber", "consume", true, events, optional.None[string]()),
	)
}

func TestRetry(t *testing.T) {
	retriesPerCall := 2
	in.Run(t,
		in.WithLanguages("java", "go"),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "firstCall"}, func(t testing.TB, resp in.Obj) {}),
		in.Call("publisher", "publishOneToTopic2", map[string]any{"haystack": "secondCall"}, func(t testing.TB, resp in.Obj) {}),

		in.Sleep(time.Second*7),

		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("firstCall")),
		checkConsumed("subscriber", "consumeButFailAndRetry", false, retriesPerCall+1, optional.Some("secondCall")),
	)
}

func TestExternalPublishRuntimeCheck(t *testing.T) {
	// No java as there is no API for this
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		in.ExpectError(
			in.Call("subscriber", "publishToExternalModule", in.Obj{}, func(t testing.TB, resp in.Obj) {}),
			"can not publish to another module's topic",
		),
	)
}

func checkConsumed(module, verb string, success bool, count int, needle optional.Option[string]) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		if needle, ok := needle.Get(); ok {
			in.Infof("Checking for %v call(s) with needle %v", count, needle)
		} else {
			in.Infof("Checking for %v call(s)", count)
		}
		resp, err := ic.Timeline.GetTimeline(ic.Context, connect.NewRequest(&timelinepb.GetTimelineRequest{
			Limit: 100000,
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
				{
					Filter: &timelinepb.GetTimelineRequest_Filter_Module{
						Module: &timelinepb.GetTimelineRequest_ModuleFilter{
							Module: module,
							Verb:   &verb,
						},
					},
				},
			},
		}))
		assert.NoError(t, err)
		calls := slices.Filter(slices.Map(resp.Msg.Events, func(e *timelinepb.Event) *timelinepb.CallEvent {
			return e.GetCall()
		}), func(c *timelinepb.CallEvent) bool {
			if c == nil {
				return false
			}
			assert.NotEqual(t, nil, c.RequestKey, "pub sub calls need a request key")
			requestKey, err := model.ParseRequestKey(*c.RequestKey)
			assert.NoError(t, err)
			assert.Equal(t, requestKey.Payload.Origin, model.OriginPubsub, "expected pubsub origin")
			if needle, ok := needle.Get(); ok && !strings.Contains(c.Request, needle) {
				return false
			}
			return true
		})
		successfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error == nil
		})
		unsuccessfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error != nil
		})
		if success {
			assert.Equal(t, count, len(successfulCalls), "expected %v successful calls (failed calls: %v)", count, len(unsuccessfulCalls))
		} else {
			assert.Equal(t, count, len(unsuccessfulCalls), "expected %v unsuccessful calls (successful calls: %v)", count, len(successfulCalls))
		}
	}
}
