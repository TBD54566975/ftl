package main

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/controller/admin"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type CLI struct {
	Version    kong.VersionFlag `help:"Show version."`
	LogConfig  log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Endpoint   *url.URL         `default:"http://127.0.0.1:8892" help:"FTL endpoint to bind/connect to." env:"FTL_ENDPOINT"`
	ConfigFlag string           `name:"config" short:"C" help:"Path to FTL project configuration file." env:"FTL_CONFIG" placeholder:"FILE"`

	Authenticators map[string]string `help:"Authenticators to use for FTL endpoints." mapsep:"," env:"FTL_AUTHENTICATORS" placeholder:"HOST=EXE,…"`
	Insecure       bool              `help:"Skip TLS certificate verification. Caution: susceptible to machine-in-the-middle attacks."`

	Ping     pingCmd     `cmd:"" help:"Ping the FTL cluster."`
	Status   statusCmd   `cmd:"" help:"Show FTL status."`
	Init     initCmd     `cmd:"" help:"Initialize a new FTL module."`
	Dev      devCmd      `cmd:"" help:"Develop FTL modules. Will start the FTL cluster, build and deploy all modules found in the specified directories, and watch for changes."`
	PS       psCmd       `cmd:"" help:"List deployments."`
	Serve    serveCmd    `cmd:"" help:"Start the FTL server."`
	Call     callCmd     `cmd:"" help:"Call an FTL function."`
	Update   updateCmd   `cmd:"" help:"Update a deployment."`
	Kill     killCmd     `cmd:"" help:"Kill a deployment."`
	Schema   schemaCmd   `cmd:"" help:"FTL schema commands."`
	Build    buildCmd    `cmd:"" help:"Build all modules found in the specified directories."`
	Deploy   deployCmd   `cmd:"" help:"Build and deploy all modules found in the specified directories."`
	Download downloadCmd `cmd:"" help:"Download a deployment."`
	Secret   secretCmd   `cmd:"" help:"Manage secrets."`
	Config   configCmd   `cmd:"" help:"Manage configuration."`

	// Specify the 1Password vault to access secrets from.
	Vault string `name:"opvault" help:"1Password vault to be used for secrets. The name of the 1Password item will be the <ref> and the secret will be stored in the password field." placeholder:"VAULT"`
}

var cli CLI

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.Configuration(kongtoml.Loader, ".ftl.toml", "~/.ftl.toml"),
		kong.ShortUsageOnError(),
		kong.HelpOptions{Compact: true, WrapUpperBound: 80},
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			node, ok := parent.(*kong.Command)
			if !ok {
				return nil
			}
			return &kong.Group{Key: node.Name, Title: "Command flags:"}
		}),
		kong.Vars{
			"version": ftl.Version,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
			"numcpu":  strconv.Itoa(runtime.NumCPU()),
		},
	)

	rpc.InitialiseClients(cli.Authenticators, cli.Insecure)

	// Set some envars for child processes.
	os.Setenv("LOG_LEVEL", cli.LogConfig.Level.String())

	ctx, cancel := context.WithCancel(context.Background())

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	if cli.Insecure {
		logger.Warnf("--insecure skips TLS certificate verification")
	}

	configPath := cli.ConfigFlag
	if configPath == "" {
		var ok bool
		configPath, ok = projectconfig.DefaultConfigPath().Get()
		if !ok {
			kctx.Fatalf("could not determine default config path, either place an ftl-project.toml file in the root of your project, use --config=FILE, or set the FTL_CONFIG envar")
		}
	}

	os.Setenv("FTL_CONFIG", configPath)

	config, err := projectconfig.Load(ctx, configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		kctx.Fatalf(err.Error())
	}
	kctx.Bind(config)

	sr := cf.ProjectConfigResolver[cf.Secrets]{Config: configPath}
	cr := cf.ProjectConfigResolver[cf.Configuration]{Config: configPath}
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

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Debugf("FTL terminating with signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	adminServiceClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, cli.Endpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, adminServiceClient)
	adminClient, err := admin.NewClient(ctx, adminServiceClient, cli.Endpoint)
	kctx.FatalIfErrorf(err)
	kctx.BindTo(adminClient, (*admin.Client)(nil))

	controllerServiceClient := rpc.Dial(ftlv1connect.NewControllerServiceClient, cli.Endpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, controllerServiceClient)
	kctx.BindTo(controllerServiceClient, (*ftlv1connect.ControllerServiceClient)(nil))

	verbServiceClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, cli.Endpoint.String(), log.Error)
	ctx = rpc.ContextWithClient(ctx, verbServiceClient)
	kctx.BindTo(verbServiceClient, (*ftlv1connect.VerbServiceClient)(nil))

	kctx.Bind(cli.Endpoint)
	kctx.BindTo(ctx, (*context.Context)(nil))

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}
