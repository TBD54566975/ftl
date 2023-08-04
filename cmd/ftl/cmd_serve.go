package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/controller"
)

type serveCmd struct {
	controller.Config
}

func (s *serveCmd) Run(ctx context.Context) error {
	return controller.Start(ctx, s.Config)
}
