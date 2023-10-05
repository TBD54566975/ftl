package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/model"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type killCmd struct {
	Deployment model.DeploymentName `arg:"" help:"Deployment to kill."`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{DeploymentName: k.Deployment.String()}))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
