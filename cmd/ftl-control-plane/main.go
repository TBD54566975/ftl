package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/controlplane"
)

var cli struct {
	LogConfig          log.Config          `embed:"" prefix:"log-"`
	ControlPlaneConfig controlplane.Config `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
	)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := controlplane.Start(ctx, cli.ControlPlaneConfig)
	kctx.FatalIfErrorf(err)
}
