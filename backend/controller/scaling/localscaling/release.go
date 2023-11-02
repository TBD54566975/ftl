//go:build release

package localscaling

import (
	"archive/zip"
	"bytes"
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"sync"

	"github.com/TBD54566975/ftl/internal"
)

//go:embed template.zip
var archive []byte

var templateDirOnce sync.Once

func templateDir(ctx context.Context) string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	cacheDir = filepath.Join(cacheDir, "ftl-runner-template")
	templateDirOnce.Do(func() {
		_ = os.RemoveAll(cacheDir)
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			panic(err)
		}
		err = internal.UnzipDir(zr, cacheDir)
		if err != nil {
			panic(err)
		}
	})
	return cacheDir
}
