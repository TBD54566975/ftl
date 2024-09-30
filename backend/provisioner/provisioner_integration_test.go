//go:build integration

package provisioner_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
	"github.com/alecthomas/assert/v2"
)

func TestDeploymentThroughProvisioner(t *testing.T) {
	in.Run(t,
		in.WithProvisioner(),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
	)
}
