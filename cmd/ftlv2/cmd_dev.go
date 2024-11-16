package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/v2/backend/schemaservice"
)

type devCmd struct {
	Bind *url.URL `help:"The address to bind the schema service to." default:"http://127.0.0.1:9992"`
}

func (c *devCmd) Run(ctx context.Context, logger *log.Logger) error {
	logger = logger.Scope("dev")
	logger.Infof("Starting dev server on %s", c.Bind)
	err := schemaservice.Start(ctx, schemaservice.Config{
		Bind: c.Bind,
	})
	if err != nil {
		return fmt.Errorf("failed to start dev server: %w", err)
	}
	return nil
}
