//go:build integration

package console_test

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"

	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1"
	in "github.com/TBD54566975/ftl/internal/integration"
)

// GetModules calls console service GetModules and returns the response.
//
// This action is defined here vs. actions.go because it is only used in this test file.
func GetModules(onResponse func(t testing.TB, resp *connect.Response[pbconsole.GetModulesResponse])) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("GetModules")
		modules, err := ic.Console.GetModules(ic.Context, &connect.Request[pbconsole.GetModulesRequest]{})
		assert.NoError(t, err)
		onResponse(t, modules)
	}
}

func TestConsoleGetModules(t *testing.T) {
	in.Run(t,
		in.CopyModule("console"),
		in.Deploy("console"),
		GetModules(func(t testing.TB, resp *connect.Response[pbconsole.GetModulesResponse]) {
			assert.Equal(t, 1, len(resp.Msg.Modules))
			assert.Equal(t, "console", resp.Msg.Modules[0].Name)
		}),
	)
}

// StreamModules calls console service GetModules and returns the response.
//
// This action is defined here vs. actions.go because it is only used in this test file.
func StreamModules(onResponse func(t testing.TB, resp *connect.ServerStreamForClient[pbconsole.StreamModulesResponse])) in.Action {
	return func(t testing.TB, ic in.TestContext) {
		in.Infof("StreamModules")
		modules, err := ic.Console.StreamModules(ic.Context, &connect.Request[pbconsole.StreamModulesRequest]{})
		assert.NoError(t, err)
		onResponse(t, modules)
	}
}

func TestConsoleStreamModules(t *testing.T) {
	in.Run(t,
		in.CopyModule("console"),
		in.Deploy("console"),
		StreamModules(func(t testing.TB, stream *connect.ServerStreamForClient[pbconsole.StreamModulesResponse]) {
			for stream.Receive() {
				assert.Equal(t, 2, len(stream.Msg().Modules))
				assert.Equal(t, "console", stream.Msg().Modules[0].Name)
				assert.Equal(t, "builtin", stream.Msg().Modules[1].Name)
				break
			}
		}),
	)
}
