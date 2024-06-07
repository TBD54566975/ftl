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
	External    []string `help:"Directories for libraries that require FTL module stubs." type:"existingdir" optional:""`
}

func (b *buildCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	if len(b.Dirs) == 0 && len(b.External) == 0 {
		b.Dirs = projConfig.AbsModuleDirsOrDefault()
		b.External = projConfig.ExternalDirs
	}
	if len(b.Dirs) == 0 && len(b.External) == 0 {
		return errors.New("no directories specified")
	}
	engine, err := buildengine.New(ctx, client, b.Dirs, b.External, buildengine.Parallelism(b.Parallelism))
	if err != nil {
		return err
	}
	return engine.Build(ctx)
}
