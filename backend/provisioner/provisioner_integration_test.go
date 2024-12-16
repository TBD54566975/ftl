//go:build integration

package provisioner_test

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestDeploymentThroughDevProvisionerCreatePostgresDB(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("./ftl-project.toml"),
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
