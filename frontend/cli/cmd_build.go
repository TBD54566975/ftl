package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type buildCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
	BuildEnv    []string `help:"Environment variables to set for the build."`
}

func (b *buildCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient, projConfig projectconfig.Config) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}
	// use the cli endpoint to create the bind allocator, but leave the first port unused as it is meant to be reserved by a controller
	bindAllocator, err := bind.NewBindAllocator(cli.Endpoint)
	if err != nil {
		return fmt.Errorf("could not create bind allocator: %w", err)
	}
	_ = bindAllocator.Next()

	engine, err := buildengine.New(ctx, client, projConfig, b.Dirs, bindAllocator, buildengine.BuildEnv(b.BuildEnv), buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}
