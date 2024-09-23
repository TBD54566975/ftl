package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type deployCmd struct {
	Replicas int32    `short:"n" help:"Number of replicas to deploy." default:"1"`
	NoWait   bool     `help:"Do not wait for deployment to complete." default:"false"`
	Build    buildCmd `embed:""`
}

func (d *deployCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	// TODO: make a better default and make it an parameter with a default
	initialBind, err := url.Parse("http://192.0.0.1:47231")
	if err != nil {
		return fmt.Errorf("failed to parse initial bind: %w", err)
	}
	bindAllocator, err := bind.NewBindAllocator(initialBind)
	if err != nil {
		return fmt.Errorf("failed to create bind allocator: %w", err)
	}
	engine, err := buildengine.New(ctx, client, bindAllocator, projConfig.Root(), d.Build.Dirs, buildengine.BuildEnv(d.Build.BuildEnv), buildengine.Parallelism(d.Build.Parallelism))
	if err != nil {
		return err
	}
	return engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait)
}
