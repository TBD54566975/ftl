package main

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
)

type buildCmd struct {
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)

	startTime := time.Now()

	module, err := buildengine.LoadModule(b.ModuleDir)
	if err != nil {
		return err
	}
	logger.Infof("Building %s module '%s'", module.Language, module.Module)

	err = buildengine.Build(ctx, module)
	if err != nil {
		return err
	}

	logger.Infof("Successfully built module '%s' in %.2fs", module.Module, time.Since(startTime).Seconds())
	return nil
}
