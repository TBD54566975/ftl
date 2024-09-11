//go:build integration

package k8sscaling_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestKubeScaling(t *testing.T) {
	in.Run(t,
		in.WithKubernetes(),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
	)
}
