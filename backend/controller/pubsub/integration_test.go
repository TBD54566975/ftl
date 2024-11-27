//go:build integration

package pubsub

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/async"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/alecthomas/assert/v2"
)

func TestPubSub(t *testing.T) {
	calls := 20
	events := calls * 10
	in.Run(t,
		in.WithLanguages("java", "go"),
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
		`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consume"}}.String()),
			events),
	)
}

func TestConsumptionDelay(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
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
		// Compare each, make sure none are less than 0.1s of each other
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
			WHERE ABS(EXTRACT(EPOCH FROM (event_times.created_at - async_call_times.created_at))) < 0.1;
		`, 0),
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
		in.Call("publisher", "publishOneToTopic2", in.Obj{}, func(t testing.TB, resp in.Obj) {}),

		in.Sleep(time.Second*6),

		// check that there are the right amount of failed async calls to the verb
		in.QueryRow("ftl",
			fmt.Sprintf(`
				SELECT COUNT(*)
				FROM async_calls
				WHERE
					state = 'error'
					AND verb = 'subscriber.consumeButFailAndRetry'
					AND catching = false
					AND origin = '%s'
		`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
			1+retriesPerCall),

		// check that there is one failed attempt to catch (we purposely fail the first one)
		in.QueryRow("ftl",
			fmt.Sprintf(`
			SELECT COUNT(*)
			FROM async_calls
			WHERE
				state = 'error'
				AND verb = 'subscriber.consumeButFailAndRetry'
				AND error LIKE '%%subscriber.catch %%'
				AND error LIKE '%%catching error%%'
				AND catching = true
				AND origin = '%s'
	`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
			1),

		// check that there is one successful attempt to catch (we succeed the second one as long as we receive the correct error in the request)
		in.QueryRow("ftl",
			fmt.Sprintf(`
		SELECT COUNT(*)
		FROM async_calls
		WHERE
			state = 'success'
			AND verb = 'subscriber.consumeButFailAndRetry'
			AND error IS NULL
			AND catching = true
			AND origin = '%s'
`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndRetry"}}.String()),
			1),

		// check that there was one successful attempt to catchAny
		in.QueryRow("ftl",
			fmt.Sprintf(`
		SELECT COUNT(*)
		FROM async_calls
		WHERE
			state = 'success'
			AND verb = 'subscriber.consumeButFailAndCatchAny'
			AND error IS NULL
			AND catching = true
			AND origin = '%s'
`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "consumeButFailAndCatchAny"}}.String()),
			1),
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

func TestLeaseFailure(t *testing.T) {
	t.Skip()
	logFilePath := filepath.Join(t.TempDir(), "pubsub.log")
	t.Setenv("TEST_LOG_FILE", logFilePath)

	in.Run(t,
		in.CopyModule("slow"),
		in.Deploy("slow"),

		// publish 2 events, with the first taking a long time to consume
		in.Call("slow", "publish", in.Obj{
			"durations": []int{20, 1},
		}, func(t testing.TB, resp in.Obj) {}),

		// while it is consuming the first event, force delete the lease in the db
		in.QueryRow("ftl", `
			WITH deleted_rows AS (
				DELETE FROM leases WHERE id = (
					SELECT lease_id FROM async_calls WHERE verb = 'slow.consume'
				)
				RETURNING *
			)
			SELECT COUNT(*) FROM deleted_rows;
		`, 1),

		in.Sleep(time.Second*7),

		// confirm that the first event failed and the second event succeeded,
		in.QueryRow("ftl", `SELECT state, error FROM async_calls WHERE verb = 'slow.consume' ORDER BY created_at`, "error", "async call lease expired"),
		in.QueryRow("ftl", `SELECT state, error FROM async_calls WHERE verb = 'slow.consume' ORDER BY created_at OFFSET 1`, "success", nil),

		// confirm that the first call did not keep executing for too long after the lease was expired
		in.IfLanguage("go",
			in.ExpectError(
				in.FileContains(logFilePath, "slept for 5s"),
				"Haystack does not contain needle",
			),
		),
	)
}

// TestIdlePerformance tests that async calls are created quickly after an event is published
func TestIdlePerformance(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go"),
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish a number of events with a delay between each
		in.Repeat(5, func(t testing.TB, ic in.TestContext) {
			in.Call("publisher", "publishOne", in.Obj{}, func(t testing.TB, resp in.Obj) {})(t, ic)
			in.Sleep(time.Millisecond*1200)(t, ic)
		}),

		// compare publication date and consumption date of each event
		in.ExpectError(func(t testing.TB, ic in.TestContext) {
			badResult := in.GetRow(t, ic, "ftl", `
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
				SELECT ABS(EXTRACT(EPOCH FROM (event_times.created_at - async_call_times.created_at)))
				FROM event_times
				JOIN async_call_times ON event_times.row_num = async_call_times.row_num
				WHERE ABS(EXTRACT(EPOCH FROM (event_times.created_at - async_call_times.created_at))) > 0.2
				LIMIT 1;
			`, 1)
			assert.True(t, false, "async calls should be created quickly after an event is published, but it took %vs", badResult[0])
		}, "sql: no rows in result set"), // no rows found means that all events were consumed quickly
	)
}
