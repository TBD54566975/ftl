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
	gitRoot, ok := internal.GitRoot("").Get()
	if !ok {
		// If GitRoot encounters an error, it will fail to find the correct dir.
		// This line preserves the original behavior to prevent a regression, but
		// it is still not the desired outcome. More thinking needed.
		gitRoot = ""
	}
	templateDirOnce.Do(func() {
		// TODO: Figure out how to make maven build offline
		err := exec.Command(ctx, log.Debug, gitRoot, "just", "build-kt-runtime").RunBuffered(ctx)
		if err != nil {
			panic(err)
		}
	})
	return filepath.Join(gitRoot, "build/template")
}
