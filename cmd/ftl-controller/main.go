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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	leasesdal "github.com/TBD54566975/ftl/backend/controller/leases/dal"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	cf "github.com/TBD54566975/ftl/common/configuration"
	cfdal "github.com/TBD54566975/ftl/common/configuration/dal"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
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
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, "ftl-controller", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	// The FTL controller currently only supports DB as a configuration provider/resolver.
	conn, err := pgxpool.New(ctx, cli.ControllerConfig.DSN)
	kctx.FatalIfErrorf(err)
	ldal := leasesdal.New(conn)
	dal, err := dal.New(ctx, conn)
	kctx.FatalIfErrorf(err)

	configDal, err := cfdal.New(ctx, conn)
	kctx.FatalIfErrorf(err)
	configProviders := []cf.Provider[cf.Configuration]{cf.NewDBConfigProvider(configDal)}
	configResolver := cf.NewDBConfigResolver(configDal)
	cm, err := cf.New[cf.Configuration](ctx, configResolver, configProviders)
	kctx.FatalIfErrorf(err)

	ctx = cf.ContextWithConfig(ctx, cm)

	// The FTL controller currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	kctx.FatalIfErrorf(err)
	secretsResolver := cf.NewASM(ctx, secretsmanager.NewFromConfig(awsConfig), cli.ControllerConfig.Advertise, ldal)
	secretsProviders := []cf.Provider[cf.Secrets]{secretsResolver}
	sm, err := cf.New[cf.Secrets](ctx, secretsResolver, secretsProviders)
	kctx.FatalIfErrorf(err)
	ctx = cf.ContextWithSecrets(ctx, sm)

	err = controller.Start(ctx, cli.ControllerConfig, scaling.NewK8sScaling(), dal)
	kctx.FatalIfErrorf(err)
}
