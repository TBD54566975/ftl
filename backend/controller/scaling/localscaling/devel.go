//go:build !release

package localscaling

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/internal"
)

var templateDirOnce sync.Once

func templateDir(ctx context.Context) string {
	templateDirOnce.Do(func() {
		cmd := exec.Command(ctx, log.Info, internal.FTLSourceRoot(), "bit", "build/template/ftl/jars/ftl-runtime.jar")
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	})
	return filepath.Join(internal.FTLSourceRoot(), "build/template")
}
