package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/v2/backend/routingservice"
)

var cli struct {
	Version       kong.VersionFlag      `help:"Show version."`
	LogConfig     log.Config            `prefix:"log-" embed:""`
	ServiceConfig routingservice.Config `embed:""`
}

func main() {
	kctx := kong.Parse(&cli, kong.Description(`
FTL - Towards a 𝝺-calculus for large-scale systems

The RoutingService performs routing for the FTL system.
	`), kong.Vars{
		"version": ftl.Version,
	})
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err := routingservice.Start(ctx, cli.ServiceConfig)
	kctx.FatalIfErrorf(err)
}
