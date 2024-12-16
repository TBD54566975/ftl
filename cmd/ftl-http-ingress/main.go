package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/ingress"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/timeline"
	_ "github.com/block/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

var cli struct {
	Version              kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig  observability.Config `embed:"" prefix:"o11y-"`
	LogConfig            log.Config           `embed:"" prefix:"log-"`
	HTTPIngressConfig    ingress.Config       `embed:""`
	SchemaServerEndpoint *url.URL             `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineEndpoint     *url.URL             `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - HTTP Ingress`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-http-ingress", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaServerEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, schemaClient)
	timelineClient := timeline.NewClient(ctx, cli.TimelineEndpoint)
	routeManager := routing.NewVerbRouter(ctx, schemaeventsource.New(ctx, schemaClient), timelineClient)
	err = ingress.Start(ctx, cli.HTTPIngressConfig, eventSource, routeManager, timelineClient)
	kctx.FatalIfErrorf(err, "failed to start HTTP ingress")
}
