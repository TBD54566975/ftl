package main

import (
	"context"
	"errors"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type buildCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string `arg:"" help:"Base directories containing modules." type:"existingdir" optional:""`
}

func (b *buildCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	if len(b.Dirs) == 0 {
		b.Dirs = projConfig.AbsModuleDirs()
	}
	if len(b.Dirs) == 0 {
		return errors.New("no directories specified")
	}
	engine, err := buildengine.New(ctx, client, projConfig.Root(), b.Dirs, buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	return engine.Build(ctx)
}
