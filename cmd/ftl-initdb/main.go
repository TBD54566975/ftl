package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/controller/dsn"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/log"
)

var cli struct {
	log.Config
	Recreate bool   `help:"Drop and recreate the database."`
	DSN      string `help:"Postgres DSN." default:"${dsn}" env:"FTL_CONTROLLER_DSN"`
}

func main() {
	kctx := kong.Parse(&cli, kong.Vars{
		"dsn": dsn.DSN("ftl"),
	})
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.Config))
	conn, err := databasetesting.CreateForDevel(ctx, cli.DSN, cli.Recreate)
	kctx.FatalIfErrorf(err)
	conn.Close()
}
