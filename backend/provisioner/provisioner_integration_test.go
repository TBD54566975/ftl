//go:build integration

package provisioner_test

import (
	"fmt"
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/alecthomas/assert/v2"
)

func TestDeploymentThroughNoopProvisioner(t *testing.T) {
	in.Run(t,
		in.WithProvisioner(),
		in.WithProvisionerConfig(`
			default = "noop"
			plugins = [
				{ id = "noop", resources = ["postgres"] },
				{ id = "controller", resources = ["module"] },
			]
		`),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
	)
}

func TestDeploymentThrougDevProvisionerCreatePostgresDB(t *testing.T) {
	in.Run(t,
		in.WithProvisioner(),
		in.CopyModule("echo"),
		in.DropDBAction(t, "echo_echodb"),
		in.Deploy("echo"),
		func(t testing.TB, ic in.TestContext) {
			counts := in.GetRow(t, ic, "postgres", fmt.Sprintf("SELECT COUNT(*) FROM pg_catalog.pg_database WHERE datname = '%s'", "echo_echodb"), 1)
			assert.True(t, counts[0].(int64) == 1, "expected 1 database, got %d", counts[0].(int64))
		},
		in.Call("echo", "echo", "Alice", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Alice!!!", response)
		}),

		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Alice,Bob!!!", response)
		}),
	)
}
