package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type buildCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv    []string `help:"Environment variables to set for the build."`
}

func (b *buildCmd) Run(
	ctx context.Context,
	controllerClient ftlv1connect.ControllerServiceClient,
	schemaSourceFactory func() schemaeventsource.EventSource,
	projConfig projectconfig.Config,
) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	engine, err := buildengine.New(
		ctx,
		controllerClient,
		schemaSourceFactory(),
		projConfig,
		b.Dirs,
		buildengine.BuildEnv(b.BuildEnv),
		buildengine.Parallelism(b.Parallelism),
	)
	if err != nil {
		return err
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}
