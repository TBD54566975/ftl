//go:build integration

package compile_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestNonExportedDecls(t *testing.T) {
	in.Run(t,
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("notexportedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "notexportedverb"}, func(_ string) {}),
			`unsupported verb parameter type &{"echo" "EchoClient"}; verbs must have the signature func(Context, Request?, Resources...)`,
		),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	in.Run(t,
		in.CopyModule("time"),
		in.Deploy("time"),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("undefinedverb"),
		in.ExpectError(
			in.ExecWithOutput("ftl", []string{"deploy", "undefinedverb"}, func(_ string) {}),
			`unsupported verb parameter type &{"echo" "UndefinedClient"}; verbs must have the signature func(Context, Request?, Resources...)`,
		),
	)
}

func TestNonFTLTypes(t *testing.T) {
	in.Run(t,
		in.CopyModule("external"),
		in.Deploy("external"),
		in.Call("external", "echo", in.Obj{"message": "hello"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["message"])
		}),
	)
}

func TestNonStructRequestResponse(t *testing.T) {
	in.Run(t,
		in.CopyModule("two"),
		in.Deploy("two"),
		in.CopyModule("one"),
		in.Deploy("one"),
		in.Call("one", "stringToTime", "1985-04-12T23:20:50.52Z", func(t testing.TB, response time.Time) {
			assert.Equal(t, time.Date(1985, 04, 12, 23, 20, 50, 520_000_000, time.UTC), response)
		}),
	)
}
