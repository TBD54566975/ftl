//go:build integration

package tests_test

import (
	"testing"

	in "github.com/block/ftl/internal/integration"
)

func TestRuntimeReflection(t *testing.T) {
	in.Run(t,
		in.CopyModule("runtimereflection"),
		in.ExecModuleTest("runtimereflection"),
	)
}
