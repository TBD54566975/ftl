package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/backend/provisioner/scaling/k8sscaling"
	_ "github.com/block/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	ProvisionerConfig   provisioner.Config   `embed:""`
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)
	cli.ProvisionerConfig.SetDefaults()

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	controllerClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, cli.ProvisionerConfig.ControllerEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, controllerClient)

	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.ProvisionerConfig.ControllerEndpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, schemaClient)
	scaling := k8sscaling.NewK8sScaling(false, cli.ProvisionerConfig.ControllerEndpoint.String())
	err = scaling.Start(ctx)
	kctx.FatalIfErrorf(err, "error starting k8s scaling")
	registry, err := provisioner.RegistryFromConfigFile(ctx, cli.ProvisionerConfig.PluginConfigFile, controllerClient, scaling)

	kctx.FatalIfErrorf(err, "failed to create provisioner registry")

	err = provisioner.Start(ctx, cli.ProvisionerConfig, registry, controllerClient)
	kctx.FatalIfErrorf(err, "failed to start provisioner")
}
