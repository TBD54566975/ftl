package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"

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
	// TODO: make a better default and make it an parameter with a default
	initialBind, err := url.Parse("http://192.0.0.1:47231")
	if err != nil {
		return fmt.Errorf("failed to parse initial bind: %w", err)
	}
	bindAllocator, err := bind.NewBindAllocator(initialBind)
	if err != nil {
		return fmt.Errorf("failed to create bind allocator: %w", err)
	}
	engine, err := buildengine.New(ctx, client, bindAllocator, projConfig.Root(), b.Dirs, buildengine.BuildEnv(b.BuildEnv), buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}
