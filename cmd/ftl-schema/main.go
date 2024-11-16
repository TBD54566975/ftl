package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/v2/backend/schemaservice"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	LogConfig           log.Config           `prefix:"log-" embed:""`
	SchemaServiceConfig schemaservice.Config `embed:""`
}

func main() {
	kctx := kong.Parse(&cli, kong.Description(`
FTL - Towards a 𝝺-calculus for large-scale systems

The SchemaService is the central service for FTL that manages the schema of the cluster.
	`), kong.Vars{
		"version": ftl.Version,
	})
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err := schemaservice.Start(ctx, cli.SchemaServiceConfig)
	kctx.FatalIfErrorf(err)
}
