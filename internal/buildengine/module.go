package buildengine

import (
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// Module represents an FTL module in the build engine
type Module struct {
	Config moduleconfig.ModuleConfig
	// paths to deploy, relative to ModuleConfig.DeployDir
	Deploy []string

	// do not read directly, use Dependecies() instead
	dependencies []string
}

func newModule(config moduleconfig.ModuleConfig) Module {
	return Module{
		Config:       config,
		dependencies: []string{},
	}
}

func (m Module) CopyWithDeploy(files []string) Module {
	module := reflect.DeepCopy(m)
	module.Deploy = files
	return module
}

func (m Module) CopyWithDependencies(dependencies []string) Module {
	module := reflect.DeepCopy(m)
	module.dependencies = dependencies
	return module
}

// DependencyMode is an enum for dependency modes
type DependencyMode string

const (
	Raw                  DependencyMode = "Raw"
	AlwaysIncludeBuiltin DependencyMode = "AlwaysIncludingBuiltin"
)

// Dependencies returns the dependencies of the module
// Mode allows us to control how dependencies are returned.
//
// When calling language plugins, use Raw mode to ensure plugins receive the same
// dependencies that were declared.
func (m Module) Dependencies(mode DependencyMode) []string {
	dependencies := m.dependencies
	switch mode {
	case Raw:
		// leave as is
	case AlwaysIncludeBuiltin:
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
	}
	return dependencies
}
