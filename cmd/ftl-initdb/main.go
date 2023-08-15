package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/controller/databasetesting"
)

var cli struct {
	log.Config
	Recreate bool   `help:"Drop and recreate the database."`
	DSN      string `help:"Postgres DSN." default:"postgres://localhost/ftl?sslmode=disable&user=postgres&password=secret" env:"FTL_DSN"`
}

func main() {
	kctx := kong.Parse(&cli)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.Config))
	conn, err := databasetesting.CreateForDevel(ctx, cli.DSN, cli.Recreate)
	kctx.FatalIfErrorf(err)
	conn.Close()
}
