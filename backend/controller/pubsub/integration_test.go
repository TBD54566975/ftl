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
