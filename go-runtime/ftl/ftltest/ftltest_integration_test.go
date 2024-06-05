//go:build integration

package ftltest

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestModuleUnitTests(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("time"),
		in.CopyModule("wrapped"),
		in.CopyModule("verbtypes"),
		in.Build("time", "wrapped", "verbtypes"),
		in.ExecModuleTest("wrapped"),
		in.ExecModuleTest("verbtypes"),
	)
}
