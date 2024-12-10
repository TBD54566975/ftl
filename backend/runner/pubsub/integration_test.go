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
								timelinepb.EventType_EVENT_TYPE_PUBSUB_CONSUME,
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

// func TestRetry(t *testing.T) {
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

// 		// check that there was one successful attempt to catchAny
// 		in.QueryRow("ftl",
// 			fmt.Sprintf(`
// 		SELECT COUNT(*)
// 		FROM async_calls
// 		WHERE
// 			state = 'success'
// 			AND verb = 'subscriber.consumeButFailAndCatchAny'
// 			AND error IS NULL
// 			AND catching = true
// 			AND origin = '%s'
// `, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndCatchAny"}}.String()),
// 			1),
// 	)
// }

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

// func TestLeaseFailure(t *testing.T) {
// 	t.Skip()
// 	logFilePath := filepath.Join(t.TempDir(), "pubsub.log")
// 	t.Setenv("TEST_LOG_FILE", logFilePath)

// 	in.Run(t,
// 		in.CopyModule("slow"),
// 		in.Deploy("slow"),

// 		// publish 2 events, with the first taking a long time to consume
// 		in.Call("slow", "publish", in.Obj{
// 			"durations": []int{20, 1},
// 		}, func(t testing.TB, resp in.Obj) {}),

// 		// while it is consuming the first event, force delete the lease in the db
// 		in.QueryRow("ftl", `
// 			WITH deleted_rows AS (
// 				DELETE FROM leases WHERE id = (
// 					SELECT lease_id FROM async_calls WHERE verb = 'slow.consume'
// 				)
// 				RETURNING *
// 			)
// 			SELECT COUNT(*) FROM deleted_rows;
// 		`, 1),

// 		in.Sleep(time.Second*7),

// 		// confirm that the first event failed and the second event succeeded,
// 		in.QueryRow("ftl", `SELECT state, error FROM async_calls WHERE verb = 'slow.consume' ORDER BY created_at`, "error", "async call lease expired"),
// 		in.QueryRow("ftl", `SELECT state, error FROM async_calls WHERE verb = 'slow.consume' ORDER BY created_at OFFSET 1`, "success", nil),

// 		// confirm that the first call did not keep executing for too long after the lease was expired
// 		in.IfLanguage("go",
// 			in.ExpectError(
// 				in.FileContains(logFilePath, "slept for 5s"),
// 				"Haystack does not contain needle",
// 			),
// 		),
// 	)
// }
