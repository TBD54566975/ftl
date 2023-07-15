package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/controller"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
)

var version = "dev"

var cli struct {
	Version            kong.VersionFlag    `help:"Show version."`
	LogConfig          log.Config          `embed:"" prefix:"log-"`
	ControllerConfig controller.Config `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": version},
	)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := controller.Start(ctx, cli.ControllerConfig)
	kctx.FatalIfErrorf(err)
}
