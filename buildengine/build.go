package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
)

// Build a module in the given directory given the schema and module config.
func Build(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx).Scope(module.Module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Building module")
	switch module.Language {
	case "go":
		return buildGo(ctx, sch, module)

	case "kotlin":
		return buildKotlin(ctx, sch, module)

	default:
		return fmt.Errorf("unknown language %q", module.Language)
	}
}
