//go:build integration

package ftl_test

import (
	"strings"
	"testing"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/alecthomas/assert/v2"

	"github.com/alecthomas/repr"
)

func TestLifecycle(t *testing.T) {
	in.Run(t, "",
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.Exec("ftl", "new", "go", ".", "echo"),
		in.Deploy("echo"),
		in.Call("echo", "echo", in.Obj{"name": "Bob"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestInterModuleCall(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("echo"),
		in.CopyModule("time"),
		in.Deploy("time"),
		in.Deploy("echo"),
		in.Call("echo", "echo", in.Obj{"name": "Bob"}, func(t testing.TB, response in.Obj) {
			message, ok := response["message"].(string)
			assert.True(t, ok, "message is not a string: %s", repr.String(response))
			if !strings.HasPrefix(message, "Hello, Bob!!! It is ") {
				t.Fatalf("unexpected response: %q", response)
			}
		}),
	)
}

func TestSchemaGenerate(t *testing.T) {
	in.Run(t, "",
		in.CopyDir("../schema-generate", "schema-generate"),
		in.Mkdir("build/schema-generate"),
		in.Exec("ftl", "schema", "generate", "schema-generate", "build/schema-generate"),
		in.FileContains("build/schema-generate/test.txt", "olleh"),
	)
}

func TestTypeRegistryUnitTest(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("typeregistry"),
		in.Deploy("typeregistry"),
		in.ExecModuleTest("typeregistry"),
	)
}
