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
		cmd := exec.Command(ctx, log.Debug, internal.GitRoot(""), "bit", "build/template/ftl/jars/ftl-runtime.jar")
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	})
	return filepath.Join(internal.GitRoot(""), "build/template")
}
