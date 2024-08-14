//go:build integration

package reflection_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestRuntimeReflection(t *testing.T) {
	in.Run(t,
		in.CopyModule("runtimereflection"),
		in.ExecModuleTest("runtimereflection"),
	)
}
