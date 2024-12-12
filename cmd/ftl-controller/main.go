package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	leasev1connect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag         `help:"Show version."`
	ObservabilityConfig observability.Config     `embed:"" prefix:"o11y-"`
	LogConfig           log.Config               `embed:"" prefix:"log-"`
	RegistryConfig      artefacts.RegistryConfig `embed:"" prefix:"oci-"`
	ControllerConfig    controller.Config        `embed:""`
	ConfigFlag          string                   `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
	DisableIstio        bool                     `help:"Disable Istio integration. This will prevent the creation of Istio policies to limit network traffic." env:"FTL_DISABLE_ISTIO"`
	TimelineEndpoint    *url.URL                 `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8894"`
	LeaseEndpoint       *url.URL                 `help:"Lease endpoint." env:"FTL_LEASE_ENDPOINT" default:"http://127.0.0.1:8895"`
	AdminEndpoint       *url.URL                 `help:"Admin endpoint." env:"FTL_ADMIN_ENDPOINT" default:"http://127.0.0.1:8896"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
			"dsn":     dsn.PostgresDSN("ftl"),
		},
	)
	cli.ControllerConfig.SetDefaults()

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	storage, err := artefacts.NewOCIRegistryStorage(cli.RegistryConfig)
	kctx.FatalIfErrorf(err, "failed to create OCI registry storage")

	// The FTL controller currently only supports DB as a cf provider/resolver.
	conn, err := cli.ControllerConfig.OpenDBAndInstrument()
	kctx.FatalIfErrorf(err)

	leaseClient := rpc.Dial(leasev1connect.NewLeaseServiceClient, cli.LeaseEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, leaseClient)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.ControllerConfig.Bind.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, schemaClient)

	adminClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, cli.AdminEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, adminClient)
	kctx.FatalIfErrorf(err)

	timelineClient := timeline.NewClient(ctx, cli.TimelineEndpoint)
	err = controller.Start(ctx, cli.ControllerConfig, storage, adminClient, timelineClient, conn, false)
	kctx.FatalIfErrorf(err)
}
