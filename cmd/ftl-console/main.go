package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller/console"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

var cli struct {
	Version               kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig   observability.Config `embed:"" prefix:"o11y-"`
	LogConfig             log.Config           `embed:"" prefix:"log-"`
	ConsoleConfig         console.Config       `embed:"" prefix:"console-"`
	TimelineEndpoint      *url.URL             `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	SchemaServiceEndpoint *url.URL             `help:"Schema service endpoint." env:"FTL_SCHEMA_SERVICE_ENDPOINT" default:"http://127.0.0.1:8893"`
	ControllerEndpoint    *url.URL             `help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
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

	err = console.Start(ctx, cli.ConsoleConfig, eventSource, controllerClient, timelineClient, adminClient)
	kctx.FatalIfErrorf(err, "failed to start console service")
}
