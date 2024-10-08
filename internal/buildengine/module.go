package buildengine

import (
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// Module represents an FTL module in the build engine
type Module struct {
	Config       moduleconfig.ModuleConfig
	Dependencies []string
	// paths to deploy, relative to ModuleConfig.DeployDir
	Deploy []string
}

func (m Module) CopyWithDependencies(dependencies []string) Module {
	module := reflect.DeepCopy(m)
	module.Dependencies = dependencies
	return module
}

func (m Module) CopyWithDeploy(files []string) Module {
	module := reflect.DeepCopy(m)
	module.Deploy = files
	return module
}
