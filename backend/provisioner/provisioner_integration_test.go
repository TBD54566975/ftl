//go:build integration

package provisioner_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/alecthomas/assert/v2"
)

func TestDeploymentThroughNoopProvisioner(t *testing.T) {
	in.Run(t,
		in.WithProvisioner(`
			default = "noop"
			plugins = [
				{ id = "noop", resources = ["postgres"] },
			]
		`),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
	)
}

func TestDeploymentThrougDevProvisioner(t *testing.T) {
	in.Run(t,
		in.WithProvisioner(`default = "dev"`),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
	)
}
