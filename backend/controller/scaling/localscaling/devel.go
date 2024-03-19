//go:build !release

package localscaling

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

var templateDirOnce sync.Once

func templateDir(ctx context.Context) string {
	templateDirOnce.Do(func() {
		// TODO: Figure out how to make maven build offline
		err := exec.Command(ctx, log.Debug, internal.GitRoot(""), "just", "build-kt-runtime").RunBuffered(ctx)
		if err != nil {
			panic(err)
		}
	})
	return filepath.Join(internal.GitRoot(""), "build/template")
}
