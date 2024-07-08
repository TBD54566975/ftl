package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGoModule(ctx context.Context, sch *schema.Schema, module Module, transaction ModifyFilesTransaction) error {
	if err := compile.Build(ctx, module.Config.Dir, sch, transaction); err != nil {
		return CompilerBuildError{err: fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)}
	}
	return nil
}
