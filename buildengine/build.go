package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/common/moduleconfig"
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
func LoadModule(dir string) (Module, error) {
	config, err := moduleconfig.LoadModuleConfig(dir)
	if err != nil {
		return Module{}, err
	}
	return UpdateDependencies(config)
}

// Build a module in the given directory given the schema and module config.
func Build(ctx context.Context /*schema *schema.Schema, */, module Module) error {
	switch module.Language {
	case "go":
		return buildGo(ctx, module)

	case "kotlin":
		return buildKotlin(ctx, module)

	default:
		return fmt.Errorf("unknown language %q", module.Language)
	}
}
