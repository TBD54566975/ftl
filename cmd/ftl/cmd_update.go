package main

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/common/model"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type updateCmd struct {
	Replicas   int32               `short:"n" help:"Number of replicas to deploy." default:"1"`
	Deployment model.DeploymentKey `arg:"" help:"Deployment to kill."`
}

func (u *updateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{
		DeploymentKey: u.Deployment.String(),
		MinReplicas:   u.Replicas,
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
