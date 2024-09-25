package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/terminal"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, binder terminal.KongContextBinder, cancel context.CancelFunc, client ftlv1connect.ControllerServiceClient) error {
	err := terminal.RunInteractiveConsole(ctx, k, projectConfig, binder, cancel, client)
	if err != nil {
		return fmt.Errorf("interactive console: %w", err)
	}
	return nil
}
