package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGoModule(ctx context.Context, projectRootDir string, sch *schema.Schema, module Module, transaction ModifyFilesTransaction) error {
	if err := compile.Build(ctx, projectRootDir, module.Config, sch, transaction); err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}
	return nil
}
