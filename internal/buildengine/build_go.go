package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGoModule(ctx context.Context, projectRootDir string, sch *schema.Schema, module Module, transaction ModifyFilesTransaction, buildEnv []string, devMode bool) error {
	// TODO: revisit this...
	_, _, err := compile.Build(ctx, projectRootDir, module.Config.Dir, sch, transaction, buildEnv, devMode)
	if err != nil {
		return CompilerBuildError{err: fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)}
	}
	return nil
}
