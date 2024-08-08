package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

type javaModuleContext struct {
	module Module
	*schema.Schema
}

func (e javaModuleContext) ExternalModules() []*schema.Module {
	modules := make([]*schema.Module, 0, len(e.Modules))
	for _, module := range e.Modules {
		if module.Name == e.module.Config.Module {
			continue
		}
		modules = append(modules, module)
	}
	return modules
}

func buildJavaModule(ctx context.Context, sch *schema.Schema, module Module) error {
	logger := log.FromContext(ctx)
	if err := SetPOMProperties(ctx, module.Config.Dir); err != nil {
		return fmt.Errorf("unable to update ftl.version in %s: %w", module.Config.Dir, err)
	}
	logger.Infof("Using build command '%s'", module.Config.Build)
	err := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}
