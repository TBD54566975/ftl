//go:build integration

package pubsub

import (
	"fmt"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/async"
	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/TBD54566975/ftl/internal/schema"
)

func TestPubSub(t *testing.T) {
	t.Skip("About to move away from legacy pubsub")
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

func TestRetry(t *testing.T) {
	t.Skip("About to move away from legacy pubsub")
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
	t.Skip("About to move away from legacy pubsub")
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
