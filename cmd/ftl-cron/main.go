package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

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
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
}

func main() {
	t, err := strconv.ParseInt(ftl.Timestamp, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp %q: %s", ftl.Timestamp, err))
	}
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Cron`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.Version, "timestamp": time.Unix(t, 0).Format(time.RFC3339)},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, false, "", "ftl-cron", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.CronConfig.SchemaServiceEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, schemaClient)

	rt := routing.New(ctx, schemaeventsource.New(ctx, schemaClient))
	err = cron.Start(ctx, eventSource, rt)
	kctx.FatalIfErrorf(err, "failed to start cron")
}
