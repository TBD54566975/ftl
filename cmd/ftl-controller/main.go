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
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
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
	dal, err := dal.New(ctx, conn)
	kctx.FatalIfErrorf(err)
	configProviders := []cf.Provider[cf.Configuration]{cf.NewDBConfigProvider(dal)}
	configResolver := cf.NewDBConfigResolver(dal)
	cm, err := cf.New[cf.Configuration](ctx, configResolver, configProviders)
	if err != nil {
		kctx.Fatalf(err.Error())
	}
	ctx = cf.ContextWithConfig(ctx, cm)

	// The FTL controller currently only supports AWS Secrets Manager as a secrets provider.
	awsConfig, err := config.LoadDefaultConfig(ctx)
	asmClient := secretsmanager.NewFromConfig(awsConfig)
	secretsResolver := cf.ASM{Client: *asmClient}
	secretsProviders := []cf.Provider[cf.Secrets]{cf.ASM{Client: *asmClient}}
	sm, err := cf.New[cf.Secrets](ctx, secretsResolver, secretsProviders)
	if err != nil {
		kctx.Fatalf(err.Error())
	}
	ctx = cf.ContextWithSecrets(ctx, sm)

	// Load the project config.
	configPath := cli.ConfigFlag
	if configPath == "" {
		var ok bool
		configPath, ok = projectconfig.DefaultConfigPath().Get()
		if !ok {
			kctx.Fatalf("could not determine default config path, either place an ftl-project.toml file in the root of your project, use --config=FILE, or set the FTL_CONFIG envar")
		}
	}
	os.Setenv("FTL_CONFIG", configPath)
	projConfig, err := projectconfig.Load(ctx, configPath)
	kctx.FatalIfErrorf(err)

	err = controller.Start(ctx, cli.ControllerConfig, projConfig, scaling.NewK8sScaling())
	kctx.FatalIfErrorf(err)
}
