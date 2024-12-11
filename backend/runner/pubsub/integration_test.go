//go:build integration

package pubsub

import (
	"testing"
	"time"

	"connectrpc.com/connect"
	timelinepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
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
		checkSuccessfullyConsumed("subscriber", "consume", events),
	)
}

func checkSuccessfullyConsumed(module, verb string, count int) in.Action {
	return func(t testing.TB, ic in.TestContext) {
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
			return true
		})
		successfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error == nil
		})
		unsuccessfulCalls := slices.Filter(calls, func(call *timelinepb.CallEvent) bool {
			return call.Error != nil
		})
		assert.Equal(t, count, len(successfulCalls), "expected %v successful calls, the following calls failed:\n%v", count, unsuccessfulCalls)
		assert.NoError(t, err)
	}
}

func TestRetry(t *testing.T) {
	t.Skip("Not implemented for new pubsub yet")
	// 	retriesPerCall := 2
	// 	in.Run(t,
	// 		in.WithLanguages("java", "go"),
	// 		in.CopyModule("publisher"),
	// 		in.CopyModule("subscriber"),
	// 		in.Deploy("publisher"),
	// 		in.Deploy("subscriber"),

	// 		// publish events
	// 		in.Call("publisher", "publishOneToTopic2", in.Obj{}, func(t testing.TB, resp in.Obj) {}),

	// 		in.Sleep(time.Second*6),

	// 		// check that there are the right amount of failed async calls to the verb
	// 		in.QueryRow("ftl",
	// 			fmt.Sprintf(`
	// 				SELECT COUNT(*)
	// 				FROM async_calls
	// 				WHERE
	// 					state = 'error'
	// 					AND verb = 'subscriber.consumeButFailAndRetry'
	// 					AND catching = false
	// 					AND origin = '%s'
	// 		`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
	// 			1+retriesPerCall),

	// 		// check that there is one failed attempt to catch (we purposely fail the first one)
	// 		in.QueryRow("ftl",
	// 			fmt.Sprintf(`
	// 			SELECT COUNT(*)
	// 			FROM async_calls
	// 			WHERE
	// 				state = 'error'
	// 				AND verb = 'subscriber.consumeButFailAndRetry'
	// 				AND error LIKE '%%subscriber.catch %%'
	// 				AND error LIKE '%%catching error%%'
	// 				AND catching = true
	// 				AND origin = '%s'
	// 	`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
	// 			1),

	// 		// check that there is one successful attempt to catch (we succeed the second one as long as we receive the correct error in the request)
	// 		in.QueryRow("ftl",
	// 			fmt.Sprintf(`
	// 		SELECT COUNT(*)
	// 		FROM async_calls
	// 		WHERE
	// 			state = 'success'
	// 			AND verb = 'subscriber.consumeButFailAndRetry'
	// 			AND error IS NULL
	// 			AND catching = true
	// 			AND origin = '%s'
	// `, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
	// 			1),

	//	// check that there was one successful attempt to catchAny
	//	in.QueryRow("ftl",
	//		fmt.Sprintf(`
	//	SELECT COUNT(*)
	//	FROM async_calls
	//	WHERE
	//		state = 'success'
	//		AND verb = 'subscriber.consumeButFailAndCatchAny'
	//		AND error IS NULL
	//		AND catching = true
	//		AND origin = '%s'
	//
	// `, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndCatchAny"}}.String()),
	//
	//			1),
	//	)
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
