package buildengine

import (
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// Module represents an FTL module in the build engine
type Module struct {
	Config       moduleconfig.ModuleConfig
	Dependencies []string
}

func (m Module) CopyWithDependencies(dependencies []string) Module {
	module := reflect.DeepCopy(m)
	module.Dependencies = dependencies
	return module
}

// LoadModule loads a module from the given directory.
func LoadModule(dir string) (Module, error) {
	config, err := moduleconfig.LoadModuleConfig(dir)
	if err != nil {
		return Module{}, err
	}
	module := Module{
		Config:       config,
		Dependencies: []string{},
	}
	return module, nil
}
