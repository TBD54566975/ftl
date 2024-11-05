package main

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/download"
	"github.com/TBD54566975/ftl/internal/model"
)

type downloadCmd struct {
	Dest       string                   `short:"d" help:"Destination directory." default:"."`
	Deployment model.DeploymentKey      `help:"Deployment to download." arg:""`
	Registry   artefacts.RegistryConfig `embed:""`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	err := download.Artefacts(ctx, client, d.Deployment, d.Dest, d.Registry)
	if err != nil {
		return fmt.Errorf("failed to download artefacts: %w", err)
	}
	return nil
}
