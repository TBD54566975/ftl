//go:build integration

package pubsub

import (
	"fmt"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/schema"
	in "github.com/TBD54566975/ftl/integration"
)

func TestPubSub(t *testing.T) {
	calls := 20
	events := calls * 10
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Repeat(calls, in.Call("publisher", "publishTen", in.Obj{}, func(t testing.TB, resp in.Obj) {})),

		in.Sleep(time.Second*4),

		// check that there are the right amount of successful async calls
		in.QueryRow("ftl",
			fmt.Sprintf(`
				SELECT COUNT(*)
				FROM async_calls
				WHERE
					state = 'success'
					AND origin = '%s'
		`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "testSubscription"}}.String()),
			events),
	)
}

func TestConsumptionDelay(t *testing.T) {
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events with a small delay between each
		// pubsub should trigger its poll a few times during this period
		// each time it should continue processing each event until it reaches one that is too new to process
		func(t testing.TB, ic in.TestContext) {
			for i := 0; i < 120; i++ {
				in.Call("publisher", "publishOne", in.Obj{}, func(t testing.TB, resp in.Obj) {})(t, ic)
				time.Sleep(time.Millisecond * 25)
			}
		},

		in.Sleep(time.Second*2),

		// Get all event created ats, and all async call created ats
		// Compare each, make sure none are less than 0.2s of each other
		in.QueryRow("ftl", `
			WITH event_times AS (
				SELECT created_at, ROW_NUMBER() OVER (ORDER BY created_at) AS row_num
				FROM (
				select * from topic_events order by created_at
				) AS sub_event_times
			),
			async_call_times AS (
				SELECT created_at, ROW_NUMBER() OVER (ORDER BY created_at) AS row_num
				FROM (
				select * from async_calls ac order by created_at
				) AS sub_async_calls
			)
			SELECT COUNT(*)
			FROM event_times
			JOIN async_call_times ON event_times.row_num = async_call_times.row_num
			WHERE ABS(EXTRACT(EPOCH FROM (event_times.created_at - async_call_times.created_at))) < 0.2;
		`, 0),
	)
}

func TestRetry(t *testing.T) {
	retriesPerCall := 2
	in.Run(t,
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events
		in.Call("publisher", "publishOneToTopic2", in.Obj{}, func(t testing.TB, resp in.Obj) {}),

		in.Sleep(time.Second*6),

		// check that there are the right amount of failed async calls to the verb
		in.QueryRow("ftl",
			fmt.Sprintf(`
				SELECT COUNT(*)
				FROM async_calls
				WHERE
					state = 'error'
					AND catching = false
					AND origin = '%s'
		`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "doomedSubscription"}}.String()),
			1+retriesPerCall),

		// check that there is one failed attempt to catch (we purposely fail the first one)
		in.QueryRow("ftl",
			fmt.Sprintf(`
			SELECT COUNT(*)
			FROM async_calls
			WHERE
				state = 'error'
				AND error = 'call to verb subscriber.catch failed: catching error'
				AND catching = true
				AND origin = '%s'
	`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "doomedSubscription"}}.String()),
			1),

		// check that there is one successful attempt to catch (we succeed the second one as long as we receive the correct error in the request)
		in.QueryRow("ftl",
			fmt.Sprintf(`
		SELECT COUNT(*)
		FROM async_calls
		WHERE
			state = 'success'
			AND error IS NULL
			AND catching = true
			AND origin = '%s'
`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "doomedSubscription"}}.String()),
			1),
	)
}

func TestExternalPublishRuntimeCheck(t *testing.T) {
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
