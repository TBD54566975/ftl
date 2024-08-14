package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

func buildJavaModule(ctx context.Context, module Module) error {
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
