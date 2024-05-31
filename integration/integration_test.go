//go:build integration

package integration_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/repr"

	. "github.com/TBD54566975/ftl/integration"
)

func TestLifecycle(t *testing.T) {
	Run(t, "",
		Exec("ftl", "init", "go", ".", "echo"),
		Deploy("echo"),
		Call("echo", "echo", Obj{"name": "Bob"}, func(t testing.TB, response Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestInterModuleCall(t *testing.T) {
	Run(t, "",
		CopyModule("echo"),
		CopyModule("time"),
		Deploy("time"),
		Deploy("echo"),
		Call("echo", "echo", Obj{"name": "Bob"}, func(t testing.TB, response Obj) {
			message, ok := response["message"].(string)
			assert.True(t, ok, "message is not a string: %s", repr.String(response))
			if !strings.HasPrefix(message, "Hello, Bob!!! It is ") {
				t.Fatalf("unexpected response: %q", response)
			}
		}),
	)
}

func TestNonExportedDecls(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		Deploy("time"),
		CopyModule("echo"),
		Deploy("echo"),
		CopyModule("notexportedverb"),
		ExpectError(
			ExecWithOutput("ftl", "deploy", "notexportedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Echo, does it need to be exported?"),
	)
}

func TestUndefinedExportedDecls(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		Deploy("time"),
		CopyModule("echo"),
		Deploy("echo"),
		CopyModule("undefinedverb"),
		ExpectError(
			ExecWithOutput("ftl", "deploy", "undefinedverb"),
			"call first argument must be a function but is an unresolved reference to echo.Undefined"),
	)
}

func TestDatabase(t *testing.T) {
	Run(t, "database/ftl-project.toml",
		// deploy real module against "testdb"
		CopyModule("database"),
		CreateDBAction("database", "testdb", false),
		Deploy("database"),
		Call("database", "insert", Obj{"data": "hello"}, nil),
		QueryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		CreateDBAction("database", "testdb", true),
		ExecModuleTest("database"),
		QueryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestSchemaGenerate(t *testing.T) {
	Run(t, "",
		CopyDir("../schema-generate", "schema-generate"),
		Mkdir("build/schema-generate"),
		Exec("ftl", "schema", "generate", "schema-generate", "build/schema-generate"),
		FileContains("build/schema-generate/test.txt", "olleh"),
	)
}

func TestHttpEncodeOmitempty(t *testing.T) {
	Run(t, "",
		CopyModule("omitempty"),
		Deploy("omitempty"),
		HttpCall(http.MethodGet, "/get", JsonData(t, Obj{}), func(t testing.TB, resp *HTTPResponse) {
			assert.Equal(t, 200, resp.Status)
			_, ok := resp.JsonBody["mustset"]
			assert.True(t, ok)
			_, ok = resp.JsonBody["error"]
			assert.False(t, ok)
		}),
	)
}

func TestRuntimeReflection(t *testing.T) {
	Run(t, "",
		CopyModule("runtimereflection"),
		ExecModuleTest("runtimereflection"),
	)
}

func TestModuleUnitTests(t *testing.T) {
	Run(t, "",
		CopyModule("time"),
		CopyModule("wrapped"),
		CopyModule("verbtypes"),
		Build("time", "wrapped", "verbtypes"),
		ExecModuleTest("wrapped"),
		ExecModuleTest("verbtypes"),
	)
}
