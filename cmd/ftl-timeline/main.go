package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/timeline"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	TimelineConfig      timeline.Config      `embed:"" prefix:"timeline-"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Timeline`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-timeline", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	err = timeline.Start(ctx, cli.TimelineConfig)
	kctx.FatalIfErrorf(err, "failed to start timeline service")
}
