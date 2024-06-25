package buildengine

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func buildRustModule(ctx context.Context, _ *schema.Schema, module Module) error {
	logger := log.FromContext(ctx)

	if err := prepareFTLRoot(module); err != nil {
		return fmt.Errorf("unable to prepare FTL root for %s: %w", module.Config.Module, err)
	}

	logger.Debugf("Using build command '%s'", module.Config.Build)
	err := exec.Command(ctx, log.Debug, module.Config.Dir, "bash", "-c", module.Config.Build).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config.Module, err)
	}

	return nil
}
