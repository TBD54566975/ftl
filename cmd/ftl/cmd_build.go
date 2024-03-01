package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type buildCmd struct {
	Dirs []string `arg:"" help:"Base directories containing modules." type:"existingdir" required:""`
}

func (b *buildCmd) Run(ctx context.Context) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	engine, err := buildengine.New(ctx, client, b.Dirs...)
	if err != nil {
		return err
	}
	return engine.Build(ctx)
}
