package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGo(ctx context.Context, sch *schema.Schema, module Module) error {
	if err := compile.Build(ctx, module.Dir, sch); err != nil {
		return fmt.Errorf("failed to build module %s: %w", module.Module, err)
	}
	return nil
}
