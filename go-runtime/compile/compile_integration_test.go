//go:build integration

package compile_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/integration"
)

func TestNonExportedDecls(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("notexportedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "notexportedverb"}, func(_ string) {}),
			"call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?",
		),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("undefinedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "undefinedverb"}, func(_ string) {}),
			"call first argument must be a function but is an unresolved reference to echo.Undefined",
		),
	)
}

func TestNonFTLTypes(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("external"),
		in.Deploy("external"),
		in.Call("external", "echo", in.Obj{"message": "hello"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["message"])
		}),
	)
}
