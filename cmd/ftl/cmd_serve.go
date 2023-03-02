package main

import (
	"context"

	drivego "github.com/TBD54566975/ftl/drive-go"
)

type serveCmd struct {
	drivego.Config
}

func (r *serveCmd) Run(ctx context.Context) error {
	return drivego.Serve(ctx, r.Config)
}
