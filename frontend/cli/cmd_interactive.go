package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
	"github.com/TBD54566975/ftl/internal/terminal"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, k *kong.Kong, binder terminal.KongContextBinder, eventSource schemaeventsource.EventSource) error {
	err := terminal.RunInteractiveConsole(ctx, k, binder, eventSource)
	if err != nil {
		return fmt.Errorf("interactive console: %w", err)
	}
	return nil
}
