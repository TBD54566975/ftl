//go:build integration

package ftltest

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestModuleUnitTests(t *testing.T) {
	in.RunWithoutController(t, "wrapped/ftl-project.toml",
		in.GitInit(),
		in.CopyModule("time"),
		in.CopyModule("wrapped"),
		in.CopyModule("verbtypes"),
		in.CopyModule("pubsub"),
		in.CopyModule("subscriber"),
		in.Build("time", "wrapped", "verbtypes", "pubsub", "subscriber"),
		in.ExecModuleTest("wrapped"),
		in.ExecModuleTest("verbtypes"),
		in.ExecModuleTest("pubsub"),
		in.ExecModuleTest("subscriber"),
	)
}
