//go:build integration

package buildengine_test

import (
	"testing"

	in "github.com/block/ftl/internal/integration"
)

func TestCycleDetection(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("depcycle1"),
		in.CopyModule("depcycle2"),

		in.ExpectError(
			in.Build("depcycle1", "depcycle2"),
			`detected a module dependency cycle that impacts these modules:`,
		),
	)
}

func TestInt64BuildError(t *testing.T) {
	in.Run(t,
		in.WithTestDataDir("testdata"),
		in.CopyModule("integer"),

		in.ExpectError(
			in.Build("integer"),
			`unsupported type "int64" for field "Input"`,
			`unsupported type "int64" for field "Output"`,
		),
	)
}
