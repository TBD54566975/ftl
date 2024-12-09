package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/cron"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	CronConfig          cron.Config          `embed:""`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Cron`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-cron", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	// ctx = timeline.ContextWithClient(ctx, timeline.NewClient(ctx, cli.TimelineEndpoint))

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.CronConfig.SchemaServiceEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, schemaClient)

	routeManager := routing.NewVerbRouter(ctx, schemaeventsource.New(ctx, schemaClient))
	err = cron.Start(ctx, eventSource, routeManager)
	kctx.FatalIfErrorf(err, "failed to start cron")
}
