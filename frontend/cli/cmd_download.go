package main

import (
	"context"
	"fmt"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/model"
)

type downloadCmd struct {
	Dest       string              `short:"d" help:"Destination directory." default:"."`
	Deployment model.DeploymentKey `help:"Deployment to download." arg:""`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	err := download.Artefacts(ctx, client, d.Deployment, d.Dest)
	if err != nil {
		return fmt.Errorf("failed to download artefacts: %w", err)
	}
	return nil
}
