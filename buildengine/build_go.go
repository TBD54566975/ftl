package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGo(ctx context.Context, sch *schema.Schema, module Module) error {
	moduleConfig, ok := module.ModuleConfig()
	if !ok {
		return fmt.Errorf("module %s is not a FTL module", module.Key())
	}
	if err := compile.Build(ctx, moduleConfig.Dir, sch); err != nil {
		return fmt.Errorf("failed to build module %s: %w", moduleConfig.Module, err)
	}
	return nil
}
