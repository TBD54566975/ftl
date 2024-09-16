package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	leasesdal "github.com/TBD54566975/ftl/backend/controller/leases/dal"
	"github.com/TBD54566975/ftl/backend/controller/scaling/k8sscaling"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	cf "github.com/TBD54566975/ftl/internal/configuration"
	cfdal "github.com/TBD54566975/ftl/internal/configuration/dal"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/configuration/routers"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ControllerConfig    controller.Config    `embed:""`
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
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

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, false, "", "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	// The FTL controller currently only supports DB as a cf provider/resolver.
	conn, err := observability.OpenDBAndInstrument(cli.ControllerConfig.DSN)
	kctx.FatalIfErrorf(err)

	dal := leasesdal.New(conn)
	kctx.FatalIfErrorf(err)

	configDal := cfdal.New(conn)
	kctx.FatalIfErrorf(err)
	configProviders := []cf.Provider[cf.Configuration]{providers.NewDatabaseConfig(configDal)}
	configResolver := routers.NewDatabaseConfig(configDal)
	cm, err := manager.New(ctx, configResolver, configProviders)
	kctx.FatalIfErrorf(err)

	ctx = manager.ContextWithConfig(ctx, cm)

	// The FTL controller currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	asmSecretProvider := providers.NewASM(ctx, secretsmanager.NewFromConfig(awsConfig), cli.ControllerConfig.Advertise, dal)
	dbSecretResolver := routers.NewDatabaseSecrets(configDal)
	sm, err := manager.New[cf.Secrets](ctx, dbSecretResolver, []cf.Provider[cf.Secrets]{asmSecretProvider})
	kctx.FatalIfErrorf(err)
	ctx = manager.ContextWithSecrets(ctx, sm)

	err = controller.Start(ctx, cli.ControllerConfig, k8sscaling.NewK8sScaling(), conn, false)
	kctx.FatalIfErrorf(err)
}
