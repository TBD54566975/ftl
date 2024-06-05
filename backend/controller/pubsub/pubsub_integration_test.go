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
	in.Run(t, "",
		in.CopyModule("publisher"),
		in.CopyModule("subscriber"),
		in.Deploy("publisher"),
		in.Deploy("subscriber"),

		// publish 2 events
		in.Call("publisher", "publish", in.Obj{}, func(t testing.TB, resp in.Obj) {}),
		in.Call("publisher", "publish", in.Obj{}, func(t testing.TB, resp in.Obj) {}),

		// TODO: speed this up when we have proper pubsub implementation
		in.Sleep(time.Second*4*2),

		// check that there are 2 successful async calls
		in.QueryRow("ftl",
			fmt.Sprintf(`
		SELECT COUNT(*)
		FROM async_calls
		WHERE
			state = 'success'
			AND origin = '%s'
		`, dal.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "subscriber", Name: "test_subscription"}}.String()),
			2),
	)
}
