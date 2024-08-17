//go:build integration

package ftltest

import (
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestModuleUnitTests(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("wrapped/ftl-project.toml"),
		in.WithoutController(),
		in.GitInit(),
		in.CopyModule("time"),
		in.CopyModule("wrapped"),
		in.CopyModule("verbtypes"),
		in.CopyModule("pubsub"),
		in.CopyModule("subscriber"),
		in.CopyModule("outer"),
		in.Build("time", "wrapped", "verbtypes", "pubsub", "subscriber", "outer"),
		in.ExecModuleTest("wrapped"),
		in.ExecModuleTest("verbtypes"),
		in.ExecModuleTest("pubsub"),
		in.ExecModuleTest("subscriber"),
		in.ExecModuleTest("outer"),
	)
}
