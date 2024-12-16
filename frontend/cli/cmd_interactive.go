package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
)

type interactiveCmd struct {
}

func (i *interactiveCmd) Run(ctx context.Context, k *kong.Kong, binder terminal.KongContextBinder, eventSource func() schemaeventsource.EventSource) error {
	err := terminal.RunInteractiveConsole(ctx, k, binder, eventSource())
	if err != nil {
		return fmt.Errorf("interactive console: %w", err)
	}
	return nil
}
