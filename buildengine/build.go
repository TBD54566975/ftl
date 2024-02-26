package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
)

// A Module is a ModuleConfig with its dependencies populated.
type Module struct {
	moduleconfig.ModuleConfig
	Dependencies []string
}

// LoadModule loads a module from the given directory.
//
// A [Module] includes the module configuration as well as its dependencies
// extracted from source code.
func LoadModule(ctx context.Context, dir string) (Module, error) {
	config, err := moduleconfig.LoadModuleConfig(dir)
	if err != nil {
		return Module{}, err
	}
	return UpdateDependencies(ctx, config)
}

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
