//go:build integration

package projectconfig

import (
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestDefaultToRootWhenModuleDirsMissing(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("no-module-dirs-ftl-project.toml"),
		in.CopyModule("echo"),
		in.Exec("ftl", "build"), // Needs to be `ftl build`, not `ftl build echo`
		in.Deploy("echo"),
		in.Call("echo", "echo", in.Obj{"name": "Bob"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestConfigCmdWithoutController(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("configs-ftl-project.toml"),
		in.WithoutController(),
		in.ExecWithExpectedOutput("\"value\"\n", "ftl", "config", "get", "key"),
	)
}

func TestFindConfig(t *testing.T) {
	checkConfig := func(subdir string) in.Action {
		return func(t testing.TB, ic in.TestContext) {
			in.Infof("Running ftl config list --values")
			cmd := exec.Command(ic, log.Debug, filepath.Join(ic.WorkingDir(), subdir), "ftl", "config", "list", "--values")
			cmd.Stdout = nil
			cmd.Stderr = nil
			output, err := cmd.CombinedOutput()
			assert.NoError(t, err, "%s", output)
			assert.Equal(t, "test = \"test\"\n", string(output))
			in.Infof("Running ftl secret list --values")
			cmd = exec.Command(ic, log.Debug, filepath.Join(ic.WorkingDir(), subdir), "ftl", "secret", "list", "--values")
			cmd.Stdout = nil
			cmd.Stderr = nil
			output, err = cmd.CombinedOutput()
			assert.NoError(t, err, "%s", output)
			assert.Equal(t, "test = \"test\"\n", string(output))
		}
	}
	in.Run(t,
		in.WithoutController(),
		in.CopyModule("findconfig"),
		checkConfig("findconfig"),
		checkConfig("findconfig/subdir"),
		in.SetEnv("FTL_CONFIG", func(ic in.TestContext) string {
			return filepath.Join(ic.WorkingDir(), "findconfig", "ftl-project.toml")
		}),
		checkConfig("."),
	)
}

func TestConfigValidation(t *testing.T) {
	in.Run(t,
		in.WithFTLConfig("./validateconfig/ftl-project.toml"),
		in.CopyModule("validateconfig"),

		// Global sets never error.
		in.Chdir("validateconfig", in.Exec("ftl", "config", "set", "key", "--inline", "valueTwo")),
		in.Chdir("validateconfig", in.Exec("ftl", "config", "set", "key", "--inline", "2")),
		in.ExecWithExpectedOutput("\"2\"\n", "ftl", "config", "get", "key"),

		// No deploy yet; module sets don't error if decl isn't found.
		in.Exec("ftl", "config", "set", "validateconfig.defaultName", "--inline", "somename"),
		in.ExecWithExpectedOutput("\"somename\"\n", "ftl", "config", "get", "validateconfig.defaultName"),
		in.Exec("ftl", "config", "set", "validateconfig.count", "--inline", "1"),
		in.ExecWithExpectedOutput("\"1\"\n", "ftl", "config", "get", "validateconfig.count"),

		// This is a mismatched type, but should pass without an active deployment.
		in.Exec("ftl", "config", "set", "validateconfig.count", "--inline", "one"),

		// Deploy; validation should now be run on config sets.
		in.Deploy("validateconfig"),
		in.Exec("ftl", "config", "set", "validateconfig.defaultName", "--inline", "somenametwo"),
		in.ExecWithExpectedOutput("\"somenametwo\"\n", "ftl", "config", "get", "validateconfig.defaultName"),
		in.Exec("ftl", "config", "set", "validateconfig.count", "--inline", "2"),
		in.ExecWithExpectedOutput("\"2\"\n", "ftl", "config", "get", "validateconfig.count"),

		// With a deploy, set should fail validation on bad data type.
		in.ExecWithExpectedError("ftl: error: unknown: JSON validation failed: count has wrong type, expected Int found string",
			"ftl", "config", "set", "validateconfig.count", "--inline", "three"),

		in.ExecWithExpectedOutput("key = \"2\"\nvalidateconfig.count = \"2\"\nvalidateconfig.defaultName = \"somenametwo\"\n",
			"ftl", "config", "list", "--values"),
	)
}
