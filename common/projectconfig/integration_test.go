//go:build integration

package projectconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestCmdsCreateProjectTomlFilesIfNonexistent(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	fileName := "ftl-project-nonexistent.toml"
	configPath := filepath.Join(cwd, "testdata", "go", fileName)

	in.Run(t, fileName,
		in.CopyModule("echo"),
		in.Exec("ftl", "config", "set", "key", "--inline", "value"),
		in.FileContains(configPath, "key"),
		in.FileContains(configPath, "InZhbHVlIg"),
	)

	// The FTL config path is special-cased to use the testdata directory
	// instead of tmpDir, so we need to clean it up manually.
	fmt.Printf("Removing config file %s\n", configPath)
	err = os.Remove(configPath)
	assert.NoError(t, err)
}

func TestDefaultToRootWhenModuleDirsMissing(t *testing.T) {
	in.Run(t, "testdata/go/no-module-dirs-ftl-project.toml",
		in.CopyModule("echo"),
		in.Exec("ftl", "build"), // Needs to be `ftl build`, not `ftl build echo`
		in.Deploy("echo"),
		in.Call("echo", "echo", in.Obj{"name": "Bob"}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
	)
}

func TestConfigCmdWithoutController(t *testing.T) {
	in.RunWithoutController(t, "testdata/go/configs-ftl-project.toml",
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
