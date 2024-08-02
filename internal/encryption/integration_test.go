//go:build integration

package encryption

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	in "github.com/TBD54566975/ftl/integration"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
)

func TestLogs(t *testing.T) {
	in.RunWithEncryption(t, "",
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		in.Call[map[string]interface{}, any]("encryption", "echo", map[string]interface{}{"name": "Alice"}, nil),

		// confirm that we can read an event for that call
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Read Logs")
			resp, err := ic.Console.GetEvents(ic.Context, connect.NewRequest(&pbconsole.EventsQuery{
				Limit: 10,
			}))
			assert.NoError(t, err, "could not get events")
			fmt.Printf("Events: %v\n", resp)
			_, ok := slices.Find(resp.Msg.Events, func(e *pbconsole.Event) bool {
				call, ok := e.Entry.(*pbconsole.Event_Call)
				if !ok {
					return false
				}
				assert.Contains(t, call.Call.Request, "Alice", "request does not contain expected value")

				return true
			})
			assert.True(t, ok, "could not find event")
		},

		// confirm that we can't find that raw request string in the table
		in.QueryRow("ftl", "SELECT COUNT(*) FROM events WHERE type = 'call' LIMIT 1 ", int64(1)),
		func(t testing.TB, ic in.TestContext) {
			values := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT payload FROM events WHERE type = 'call' LIMIT 1 "), 1)
			payload, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to string")
			assert.Contains(t, string(payload), "encrypted", "raw request string should not be stored in the table")
			assert.NotContains(t, string(payload), "Alice", "raw request string should not be stored in the table")
		},
	)
}
