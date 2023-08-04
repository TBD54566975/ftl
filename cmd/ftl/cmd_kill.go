package main

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	"github.com/TBD54566975/ftl/backend/common/model"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type killCmd struct {
	Deployment model.DeploymentKey `arg:"" help:"Deployment to kill."`
}

func (k *killCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	_, err := client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{DeploymentKey: k.Deployment.String()}))
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
