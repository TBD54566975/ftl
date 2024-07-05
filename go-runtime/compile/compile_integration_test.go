//go:build integration

package compile_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestNonExportedDecls(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("notexportedverb"),
		in.ExecWithExpectedError("call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?",
			"ftl", "deploy", "notexportedverb"),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("undefinedverb"),
		in.ExecWithExpectedError("call first argument must be a function but is an unresolved reference to echo.Undefined",
			"ftl", "deploy", "undefinedverb"),
	)
}
