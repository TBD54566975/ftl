package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type deployCmd struct {
	Replicas       int32    `short:"n" help:"Number of replicas to deploy." default:"1"`
	NoWait         bool     `help:"Do not wait for deployment to complete." default:"false"`
	UseProvisioner bool     `help:"Use the ftl-provisioner to deploy the application." default:"false" hidden:"true"`
	Build          buildCmd `embed:""`
}

func (d *deployCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	var client buildengine.DeployClient
	if d.UseProvisioner {
		client = rpc.ClientFromContext[provisionerconnect.ProvisionerServiceClient](ctx)
	} else {
		client = rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	}
	engine, err := buildengine.New(ctx, client, projConfig.Root(), d.Build.Dirs, buildengine.BuildEnv(d.Build.BuildEnv), buildengine.Parallelism(d.Build.Parallelism))
	if err != nil {
		return err
	}
	return engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait)
}
