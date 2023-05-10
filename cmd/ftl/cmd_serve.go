package main

import (
	"context"

	"github.com/TBD54566975/ftl/backplane"
)

type serveCmd struct {
	backplane.Config
}

func (s *serveCmd) Run(ctx context.Context) error {
	return backplane.Start(ctx, s.Config)
}
