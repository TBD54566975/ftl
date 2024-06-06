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
	in.Run(t, "",
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
		`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "test_subscription"}}.String()),
			events),
	)
}

func TestPubSubConsumptionDelay(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish events with a small delay between each
		// pubsub should trigger its poll a few times during this period
		// each time it should continue processing each event until it reaches one that is too new to process
		func(t testing.TB, ic in.TestContext) {
			for i := 0; i < 60; i++ {
				in.Call("publisher", "publishOne", in.Obj{}, func(t testing.TB, resp in.Obj) {})(t, ic)
				time.Sleep(time.Millisecond * 50)
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
			)
		  ),
		  async_call_times AS (
			SELECT created_at, ROW_NUMBER() OVER (ORDER BY created_at) AS row_num
			FROM (
			  select * from async_calls ac order by created_at
			)
		  )
		  SELECT COUNT(*)
		  FROM event_times
		  JOIN async_call_times ON event_times.row_num = async_call_times.row_num
		  WHERE ABS(EXTRACT(EPOCH FROM (event_times.created_at - async_call_times.created_at))) < 0.2;
		`, 0),
	)
}
