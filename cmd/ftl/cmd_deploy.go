package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type deployCmd struct {
	Parallelism int      `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Replicas    int32    `short:"n" help:"Number of replicas to deploy." default:"1"`
	Dirs        []string `arg:"" help:"Base directories containing modules." type:"existingdir" required:""`
	NoWait      bool     `help:"Do not wait for deployment to complete." default:"false"`
}

func (d *deployCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	engine, err := buildengine.New(ctx, client, projConfig.Root(), d.Dirs, buildengine.Parallelism(d.Parallelism))
	if err != nil {
		return err
	}
	return engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait)
}
