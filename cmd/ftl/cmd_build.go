package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
)

type buildCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string `arg:"" help:"Base directories containing modules (defaults to modules in project config)." type:"existingdir" optional:""`
}

func (b *buildCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient, projConfig projectconfig.Config) error {
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}
	engine, err := buildengine.New(ctx, client, b.Dirs, buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	if err := engine.Build(ctx); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}
