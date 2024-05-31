package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/scaling"
	cf "github.com/TBD54566975/ftl/common/configuration"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ControllerConfig    controller.Config    `embed:""`
	ConfigFlag          []string             `name:"config" short:"C" help:"Paths to FTL project configuration files." env:"FTL_CONFIG" placeholder:"FILE[,FILE,...]"`

	// Specify the 1Password vault to access secrets from.
	Vault string `name:"op" help:"Store a secret in this 1Password vault. The name of the 1Password item will be the <ref> and the secret will be stored in the password field." group:"Provider:" xor:"secretwriter" placeholder:"VAULT"`
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

	// This is duplicating the logic in the `ftl/main.go` to resolve the current panic
	// However, this should be updated to only allow providers that are supported on the current environment
	// See https://github.com/TBD54566975/ftl/issues/1473 for more information
	sr := cf.ProjectConfigResolver[cf.Secrets]{Config: cli.ConfigFlag}
	cr := cf.ProjectConfigResolver[cf.Configuration]{Config: cli.ConfigFlag}
	kctx.BindTo(sr, (*cf.Resolver[cf.Secrets])(nil))
	kctx.BindTo(cr, (*cf.Resolver[cf.Configuration])(nil))

	// Add config manager to context.
	cm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		kctx.Fatalf(err.Error())
	}
	ctx = cf.ContextWithConfig(ctx, cm)

	// Add secrets manager to context.
	sm, err := cf.NewSecretsManager(ctx, sr, cli.Vault)
	if err != nil {
		kctx.Fatalf(err.Error())
	}
	ctx = cf.ContextWithSecrets(ctx, sm)

	err = controller.Start(ctx, cli.ControllerConfig, scaling.NewK8sScaling())
	kctx.FatalIfErrorf(err)
}
