//go:build integration

package ftl

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestFTLMap(t *testing.T) {
	in.Run(t,
		in.CopyModule("mapper"),
		in.Build("mapper"),
		in.ExecModuleTest("mapper"),
	)
}
