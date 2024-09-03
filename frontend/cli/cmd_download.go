package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/model"
)

type downloadCmd struct {
	Dest       string              `short:"d" help:"Destination directory." default:"."`
	Deployment model.DeploymentKey `help:"Deployment to download." arg:""`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	return download.Artefacts(ctx, client, d.Deployment, d.Dest)
}
