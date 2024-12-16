package compile

import (
	"context"
	"fmt"
	"maps"

	"github.com/alecthomas/types/optional"
	"github.com/block/scaffolder"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
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
