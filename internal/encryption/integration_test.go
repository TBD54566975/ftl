//go:build integration

package encryption

import (
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	in "github.com/TBD54566975/ftl/integration"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
)

func TestEncryptionForLogs(t *testing.T) {
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
		in.QueryRow("ftl", "SELECT COUNT(*) FROM events WHERE type = 'call'", int64(1)),
		func(t testing.TB, ic in.TestContext) {
			values := in.GetRow(t, ic, "ftl", "SELECT payload FROM events WHERE type = 'call' LIMIT 1", 1)
			payload, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to string")
			assert.Contains(t, string(payload), "encrypted", "raw request string should not be stored in the table")
			assert.NotContains(t, string(payload), "Alice", "raw request string should not be stored in the table")
		},
	)
}

func TestEncryptionForubSub(t *testing.T) {
	in.RunWithEncryption(t, "",
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		in.Call[map[string]interface{}, any]("encryption", "publish", map[string]interface{}{"name": "AliceInWonderland"}, nil),

		in.Sleep(4*time.Second),

		// check that the event was published with an encrypted request
		in.QueryRow("ftl", "SELECT COUNT(*) FROM topic_events", int64(1)),
		func(t testing.TB, ic in.TestContext) {
			values := in.GetRow(t, ic, "ftl", "SELECT payload FROM topic_events", 1)
			payload, ok := values[0].([]byte)
			assert.True(t, ok, "could not convert payload to string")
			assert.Contains(t, string(payload), "encrypted", "raw request string should not be stored in the table")
			assert.NotContains(t, string(payload), "AliceInWonderland", "raw request string should not be stored in the table")
		},
		validateAsyncCall("consume", "AliceInWonderland"),
	)
}

func TestEncryptionForFSM(t *testing.T) {
	in.RunWithEncryption(t, "",
		in.CopyModule("encryption"),
		in.Deploy("encryption"),
		in.Call[map[string]interface{}, any]("encryption", "beginFsm", map[string]interface{}{"name": "Rosebud"}, nil),
		in.Sleep(3*time.Second),
		in.Call[map[string]interface{}, any]("encryption", "transitionFsm", map[string]interface{}{"name": "Rosebud"}, nil),
		in.Sleep(3*time.Second),

		validateAsyncCall("created", "Rosebud"),
		validateAsyncCall("paid", "Rosebud"),
	)
}

func validateAsyncCall(verb string, sensitive string) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.QueryRow("ftl", fmt.Sprintf("SELECT COUNT(*) FROM async_calls WHERE verb = 'encryption.%s' AND state = 'success'", verb), int64(1))(t, ic)

		values := in.GetRow(t, ic, "ftl", fmt.Sprintf("SELECT request FROM async_calls WHERE verb = 'encryption.%s' AND state = 'success'", verb), 1)
		request, ok := values[0].([]byte)
		assert.True(t, ok, "could not convert payload to string")
		assert.Contains(t, string(request), "encrypted", "raw request string should not be stored in the table")
		assert.NotContains(t, string(request), sensitive, "raw request string should not be stored in the table")
	}
}
