package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/lease/v1/ftlv1connect"
	ftlv1connect2 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/timeline"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	cf "github.com/TBD54566975/ftl/internal/configuration"
	cfdal "github.com/TBD54566975/ftl/internal/configuration/dal"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
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

	kctx.FatalIfErrorf(err)

	configDal := cfdal.New(conn)
	kctx.FatalIfErrorf(err)
	configResolver := routers.NewDatabaseConfig(configDal)
	cm, err := manager.New(ctx, configResolver, providers.NewDatabaseConfig(configDal))
	kctx.FatalIfErrorf(err)

	leaseClient := rpc.Dial(ftlv1connect.NewLeaseServiceClient, cli.LeaseEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, leaseClient)
	schemaClient := rpc.Dial(ftlv1connect2.NewSchemaServiceClient, cli.ControllerConfig.Bind.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, schemaClient)
	// The FTL controller currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := providers.NewASM(secretsmanager.NewFromConfig(awsConfig))
	dbSecretResolver := routers.NewDatabaseSecrets(configDal)
	sm, err := manager.New[cf.Secrets](ctx, dbSecretResolver, asmSecretProvider)
	kctx.FatalIfErrorf(err)

	timelineClient := timeline.NewClient(ctx, cli.TimelineEndpoint)
	err = controller.Start(ctx, cli.ControllerConfig, storage, cm, sm, timelineClient, conn, false)
	kctx.FatalIfErrorf(err)
}
