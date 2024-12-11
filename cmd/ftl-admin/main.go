package main

import (
	"context"
	"net/url"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/admin"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	cf "github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

var cli struct {
	Version              kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig  observability.Config `embed:"" prefix:"o11y-"`
	LogConfig            log.Config           `embed:"" prefix:"log-"`
	AdminConfig          admin.Config         `embed:"" prefix:"admin-"`
	SchemaServerEndpoint *url.URL             `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Admin`),
		kong.UsageOnError(),
		kong.Vars{
			"version": ftl.FormattedVersion,
			"dsn":     dsn.PostgresDSN("ftl"),
		},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-admin", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	configResolver := routers.NoopRouter[cf.Configuration]{}
	cm, err := manager.New(ctx, &configResolver, providers.NewMemory[cf.Configuration]())
	kctx.FatalIfErrorf(err)

	// FTL currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := providers.NewASM(secretsmanager.NewFromConfig(awsConfig))
	dbSecretResolver := routers.NoopRouter[cf.Secrets]{}
	sm, err := manager.New[cf.Secrets](ctx, &dbSecretResolver, asmSecretProvider)
	kctx.FatalIfErrorf(err)

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.SchemaServerEndpoint.String(), log.Error)
	eventSource := schemaeventsource.New(ctx, schemaClient)

	err = admin.Start(ctx, cli.AdminConfig, cm, sm, admin.NewSchemaRetreiver(eventSource))
	kctx.FatalIfErrorf(err, "failed to start timeline service")
}
