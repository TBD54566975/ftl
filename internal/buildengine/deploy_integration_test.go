//go:build integration

package buildengine

import (
	"testing"

	in "github.com/block/ftl/internal/integration"
)

func TestDeploy(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("another"),

		// Build first to make sure the files are there.
		in.Build("another"),
		in.FileExists("/another/.ftl/main"),

		// Test that the deployment works and starts correctly
		in.Deploy("another"),
	)
}
