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
	fileName := "ftl-project-nonexistent.toml"
	in.Run(t, fileName,
		in.CopyModule("echo"),
		in.Exec("ftl", "config", "set", "key", "--inline", "value"),
	)

	// The FTL config path is special-cased to use the testdata directory
	// instead of tmpDir.
	configPath := filepath.Join("testdata", "go", fileName)

	fmt.Printf("Checking that %s exists\n", configPath)
	_, err := os.Stat(configPath)
	assert.NoError(t, err)

	fmt.Printf("Removing config file %s\n", configPath)
	err = os.Remove(configPath)
	assert.NoError(t, err)
}
