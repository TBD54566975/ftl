//go:build smoketest

package smoketest

import (
	"fmt"
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
		in.Call("echo", "echo", "Joe", func(t testing.TB, response string) {
			name := "Joe"
			fmt.Printf("name: %s\n", name)
			assert.Equal(t, "Hello, %s!!!", name, response)
		}),
		in.Exec("ftl", "--version"),
	)
}
