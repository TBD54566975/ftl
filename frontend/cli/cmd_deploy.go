package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type deployCmd struct {
	Replicas       int32    `short:"n" help:"Number of replicas to deploy." default:"1"`
	NoWait         bool     `help:"Do not wait for deployment to complete." default:"false"`
	UseProvisioner bool     `help:"Use the ftl-provisioner to deploy the application." default:"false" hidden:"true"`
	Build          buildCmd `embed:""`
}

func (d *deployCmd) Run(
	ctx context.Context,
	projConfig projectconfig.Config,
	provisionerClient provisionerconnect.ProvisionerServiceClient,
	controllerClient ftlv1connect.ControllerServiceClient,
	schemaClient ftlv1connect.SchemaServiceClient,
) error {
	var client buildengine.DeployClient
	if d.UseProvisioner {
		client = provisionerClient
	} else {
		client = controllerClient
	}

	// Cancel build engine context to ensure all language plugins are killed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	engine, err := buildengine.New(
		ctx, client, schemaClient, projConfig, d.Build.Dirs,
		buildengine.BuildEnv(d.Build.BuildEnv),
		buildengine.Parallelism(d.Build.Parallelism),
	)
	if err != nil {
		return err
	}
	return engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait)
}
