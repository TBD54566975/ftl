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

func buildSwiftModule(ctx context.Context, sch *schema.Schema, module Module) error {
	// TODO: update module package FTL dependencies

	// TODO: generate external dependencies

	if err := prepareFTLRoot(module); err != nil {
		return fmt.Errorf("unable to prepare FTL root for %s: %w", module.Config.Module, err)
	}

	goExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get go executable path: %w", err)
	}
	err = exec.Command(ctx, log.Debug,
		filepath.Dir(goExecPath),
		"./ftl-swift-compile",
		"--name", module.Config.Module,
		"--root-path", module.Config.Dir,
		"--build-cmd", module.Config.Build,
		"--deploy-path", module.Config.AbsDeployDir(),
		"--schema-filename", module.Config.Schema).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}
