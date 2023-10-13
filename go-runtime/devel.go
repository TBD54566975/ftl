//go:build !release

package goruntime

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Files is the FTL Go runtime scaffolding files.
var Files = func() fs.FS {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(strings.TrimSpace(string(out)), "go-runtime", "scaffolding")
	return os.DirFS(dir)
}()
