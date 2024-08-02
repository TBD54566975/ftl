package encryption

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	in "github.com/TBD54566975/ftl/integration"
	"github.com/alecthomas/assert/v2"
)

func TestLogs(t *testing.T) {
	in.RunWithEncryption(t, "",
		in.CopyModule("encryption"),
		in.Deploy("encryption"),

		in.Call[map[string]interface{}, any]("encryption", "echo", map[string]interface{}{"name": "Alice"}, nil),
		func(t testing.TB, ic in.TestContext) {
			in.Infof("Read Logs")

			events, err := ic.Console.GetEvents(ic.Context, connect.NewRequest(&pbconsole.EventsQuery{
				Limit: 10,
			}))
			assert.NoError(t, err, "could not get events")
			fmt.Printf("Events: %v\n", events)
		},
	)
}

// func GetModules(onResponse func(t testing.TB, resp *connect.Response[pbconsole.GetModulesResponse])) in.Action {
// 	return func(t testing.TB, ic in.TestContext) {
// 		in.Infof("GetModules")
// 		modules, err := ic.Console.GetModules(ic.Context, &connect.Request[pbconsole.GetModulesRequest]{})
// 		assert.NoError(t, err)
// 		onResponse(t, modules)
// 	}
// }

// func TestConsoleGetModules(t *testing.T) {
// 	in.Run(t, "",
// 		in.CopyModule("console"),
// 		in.Deploy("console"),
// 		GetModules(func(t testing.TB, resp *connect.Response[pbconsole.GetModulesResponse]) {
// 			assert.Equal(t, 1, len(resp.Msg.Modules))
// 			assert.Equal(t, "console", resp.Msg.Modules[0].Name)
// 		}),
// 	)
// }
