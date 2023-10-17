//go:build !release

package goruntime

import (
	"archive/zip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TBD54566975/ftl/internal"
)

// Files is the FTL Go runtime scaffolding files.
var Files = func() *zip.Reader {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(strings.TrimSpace(string(out)), "go-runtime", "scaffolding")
	w, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.Remove(w.Name()) // This is okay because the zip.Reader will keep it open.
	if err != nil {
		panic(err)
	}

	err = internal.ZipDir(dir, w.Name())
	if err != nil {
		panic(err)
	}

	info, err := w.Stat()
	if err != nil {
		panic(err)
	}
	_, _ = w.Seek(0, 0)
	zr, err := zip.NewReader(w, info.Size())
	if err != nil {
		panic(err)
	}
	return zr
}()
