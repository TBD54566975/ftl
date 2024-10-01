package buildengine

import (
	"context"

	"github.com/TBD54566975/ftl/internal/log"
)

// UpdateDependencies finds the dependencies for a module and returns a
// Module with those dependencies populated.
func UpdateDependencies(ctx context.Context, module Module, plugin LanguagePlugin) (Module, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting dependencies for %q", module.Config.Module)
	dependencies, err := plugin.GetDependencies(ctx)
	if err != nil {
		return Module{}, err
	}
	containsBuiltin := false
	for _, dep := range dependencies {
		if dep == "builtin" {
			containsBuiltin = true
			break
		}
	}
	if !containsBuiltin {
		dependencies = append(dependencies, "builtin")
	}

	out := module.CopyWithDependencies(dependencies)
	return out, nil
}
