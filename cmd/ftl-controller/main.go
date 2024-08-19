package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	cf "github.com/TBD54566975/ftl/internal/configuration"
	cfdal "github.com/TBD54566975/ftl/internal/configuration/dal"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ControllerConfig    controller.Config    `embed:""`
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project configuration file." env:"FTL_CONFIG" placeholder:"FILE"`
}

func main() {
	t, err := strconv.ParseInt(ftl.Timestamp, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp %q: %s", ftl.Timestamp, err))
	}
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.Version, "timestamp": time.Unix(t, 0).Format(time.RFC3339)},
	)
	cli.ControllerConfig.SetDefaults()

	if cli.ControllerConfig.KMSURI == nil {
		kctx.Fatalf("KMSURI is required")
	}

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, false, "", "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	// The FTL controller currently only supports DB as a configuration provider/resolver.
	conn, err := sql.Open("pgx", cli.ControllerConfig.DSN)
	kctx.FatalIfErrorf(err)

	encryptionBuilder := encryption.NewBuilder().WithKMSURI(optional.Ptr(cli.ControllerConfig.KMSURI))
	kctx.FatalIfErrorf(err)
	dal, err := dal.New(ctx, conn, encryptionBuilder)
	kctx.FatalIfErrorf(err)

	configDal, err := cfdal.New(ctx, conn)
	kctx.FatalIfErrorf(err)
	configProviders := []cf.Provider[cf.Configuration]{cf.NewDBConfigProvider(configDal)}
	configResolver := cf.NewDBConfigResolver(configDal)
	cm, err := cf.New(ctx, configResolver, configProviders)
	kctx.FatalIfErrorf(err)

	ctx = cf.ContextWithConfig(ctx, cm)

	// The FTL controller currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := cf.NewASM(ctx, secretsmanager.NewFromConfig(awsConfig), cli.ControllerConfig.Advertise, dal)
	dbSecretResolver := cf.NewDBSecretResolver(configDal)
	sm, err := cf.New[cf.Secrets](ctx, dbSecretResolver, []cf.Provider[cf.Secrets]{asmSecretProvider})
	kctx.FatalIfErrorf(err)
	ctx = cf.ContextWithSecrets(ctx, sm)

	err = controller.Start(ctx, cli.ControllerConfig, scaling.NewK8sScaling(), conn)
	kctx.FatalIfErrorf(err)
}
