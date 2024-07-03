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
	in.Run(t, "no-module-dirs-ftl-project.toml",
		in.CopyModule("echo"),
		in.Exec("ftl", "build"), // Needs to be `ftl build`, not `ftl build echo`
		in.Deploy("echo"),
		in.Call("echo", "echo", in.Obj{"name": "Bob"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestConfigCmdWithoutController(t *testing.T) {
	in.RunWithoutController(t, "configs-ftl-project.toml",
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
	in.RunWithoutController(t, "",
		in.CopyModule("findconfig"),
		checkConfig("findconfig"),
		checkConfig("findconfig/subdir"),
		in.SetEnv("FTL_CONFIG", func(ic in.TestContext) string {
			return filepath.Join(ic.WorkingDir(), "findconfig", "ftl-project.toml")
		}),
		checkConfig("."),
	)
}
