package compile

import (
	"context"
	"fmt"
	"maps"

	"github.com/alecthomas/types/optional"
	"github.com/block/scaffolder"

	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

type ExternalDeploymentContext struct {
	Name   string
	Module *schema.Module
}

func GenerateStubs(ctx context.Context, dir string, moduleSch *schema.Module, config moduleconfig.AbsModuleConfig, nativeConfig optional.Option[moduleconfig.AbsModuleConfig]) error {
	context := ExternalDeploymentContext{
		Name:   moduleSch.Name,
		Module: moduleSch,
	}

	funcs := maps.Clone(scaffoldFuncs)
	err := internal.ScaffoldZip(externalModuleTemplateFiles(), dir, context, scaffolder.Functions(funcs))
	if err != nil {
		return fmt.Errorf("failed to scaffold zip: %w", err)
	}

	if err := exec.Command(ctx, log.Debug, dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return fmt.Errorf("failed to tidy go.mod: %w", err)
	}
	return nil
}
