package main

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/controller"
)

type initDBCmd struct {
	Recreate bool   `help:"Drop and recreate the database."`
	DSN      string `help:"Postgres DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_DSN"`
}

func (c *initDBCmd) Run(ctx context.Context) error {
	conn, err := controller.CreateForDevel(ctx, c.DSN, c.Recreate)
	if err != nil {
		return errors.WithStack(err)
	}
	conn.Close()
	return nil
}
