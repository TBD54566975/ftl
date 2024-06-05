//go:build integration

package projectconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	in "github.com/TBD54566975/ftl/integration"
	"github.com/alecthomas/assert/v2"
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
