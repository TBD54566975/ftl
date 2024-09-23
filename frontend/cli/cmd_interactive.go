package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/internal/console"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, k *kong.Kong, projectConfig projectconfig.Config, binder console.KongContextBinder, cancel context.CancelFunc) error {
	err := console.RunInteractiveConsole(ctx, k, projectConfig, binder, nil, cancel)
	if err != nil {
		return fmt.Errorf("interactive console: %w", err)
	}
	return nil
}
