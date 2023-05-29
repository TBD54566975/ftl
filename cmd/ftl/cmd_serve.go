package main

import (
	"context"

	"github.com/TBD54566975/ftl/controlplane"
)

type serveCmd struct {
	controlplane.Config
}

func (s *serveCmd) Run(ctx context.Context) error {
	return controlplane.Start(ctx, s.Config)
}
