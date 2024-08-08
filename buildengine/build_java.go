package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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
	if err := writeProtosForJavaBuild(module, sch); err != nil {
		return fmt.Errorf("unable to generate external modules for %s: %w", module.Config.Module, err)
	}
	if err := prepareFTLRoot(module); err != nil {
		return fmt.Errorf("unable to prepare FTL root for %s: %w", module.Config.Module, err)
	}

	logger.Infof("Using build command '%s'", module.Config.Build)
	err := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}

func writeProtosForJavaBuild(module Module, sch *schema.Schema) error {

	// We rely on the Java build infrastructure to actually generate the Java source files
	// We just write out the relevant protos
	// This allows the JVM build infra to make more informed decisions about how to build
	// e.g. it can examine the class path to decide on which types already exist vs which need to be generated

	modPath := filepath.Join(module.Config.Dir, "src", "main", "ftl-schema")
	err := os.MkdirAll(modPath, 0744)
	if err != nil {
		return err
	}

	for _, mod := range sch.Modules {
		if mod.Name == module.Config.Module {
			continue
		}
		data, err := schema.ModuleToBytes(mod)
		if err != nil {
			return fmt.Errorf("failed to export module schema for module %s %w", mod.Name, err)
		}
		err = os.WriteFile(filepath.Join(modPath, mod.Name+".pb"), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
