package main

import (
	"context"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/model"
)

type updateCmd struct {
	Replicas   int32               `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment model.DeploymentKey `arg:"" help:"Deployment to update."`
}

func (u *updateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{
		DeploymentKey: u.Deployment.String(),
		MinReplicas:   u.Replicas,
	}))
	if err != nil {
		return err
	}
	return nil
}
