package buildengine

import (
	"github.com/TBD54566975/ftl/internal/moduleconfig"
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
