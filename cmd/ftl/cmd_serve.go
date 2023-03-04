package main

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/torres"
)

type serveCmd struct {
	Dir string `arg:"" help:"Path to an FTL module."`
}

func (r *serveCmd) Run(ctx context.Context) error {
	engineer, err := torres.New(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	err = engineer.Manage(ctx, r.Dir)
	if err != nil {
		return errors.WithStack(err)
	}
	return engineer.Wait()
}
