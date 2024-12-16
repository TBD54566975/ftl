package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/console"
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
	Version               kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig   observability.Config `embed:"" prefix:"o11y-"`
	LogConfig             log.Config           `embed:"" prefix:"log-"`
	ConsoleConfig         console.Config       `embed:"" prefix:"console-"`
	TimelineEndpoint      *url.URL             `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	SchemaServiceEndpoint *url.URL             `help:"Schema service endpoint." env:"FTL_SCHEMA_SERVICE_ENDPOINT" default:"http://127.0.0.1:8893"`
	ControllerEndpoint    *url.URL             `help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	VerbServiceEndpoint   *url.URL             `help:"Verb service endpoint." env:"FTL_VERB_SERVICE_ENDPOINT" default:"http://127.0.0.1:8895"`
	AdminEndpoint         *url.URL             `help:"Admin endpoint." env:"FTL_ADMIN_ENDPOINT" default:"http://127.0.0.1:8896"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Console`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-console", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	timelineClient := timeline.NewClient(ctx, cli.TimelineEndpoint)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaServiceEndpoint.String(), log.Error)
	controllerClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, cli.ControllerEndpoint.String(), log.Error)
	adminClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, schemaClient)

	routeManager := routing.NewVerbRouter(ctx, schemaeventsource.New(ctx, schemaClient), timelineClient)

	err = console.Start(ctx, cli.ConsoleConfig, eventSource, controllerClient, timelineClient, adminClient, routeManager)
	kctx.FatalIfErrorf(err, "failed to start console service")
}
