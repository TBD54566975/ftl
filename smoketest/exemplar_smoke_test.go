//go:build smoketest

package smoketest

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestSmokeEcho(t *testing.T) {
	in.Run(t,
		in.WithKubernetes(),
		in.WithTestDataDir("."),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.Exec("ftl", "--version"),
	)
}
