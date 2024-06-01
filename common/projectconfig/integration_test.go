//go:build integration

package projectconfig

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestCmdsCreateProjectTomlFilesIfNonexistent(t *testing.T) {
	in.Run(t, "",
		in.CopyModule("echo"),
		in.Exec("ftl", "config", "list", "--config", "ftl-project-nonexistent.toml"),
		in.FileExists("ftl-project-nonexistent.toml"),
	)
}
