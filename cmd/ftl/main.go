package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/kong-toml"
	"github.com/bufbuild/connect-go"

	_ "github.com/TBD54566975/ftl/backend/common/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/rpc"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var version = "dev"

type CLI struct {
	Version   kong.VersionFlag `help:"Show version."`
	Config    kong.ConfigFlag  `help:"Load configuration from TOML file." placeholder:"FILE"`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Endpoint  *url.URL         `default:"http://127.0.0.1:8892" help:"FTL endpoint to bind/connect to." env:"FTL_ENDPOINT"`

	Authenticators map[string]string `help:"Authenticators to use for FTL endpoints." mapsep:"," env:"FTL_AUTHENTICATORS" placeholder:"HOST=EXE,…"`

	Status   statusCmd   `cmd:"" help:"Show FTL status."`
	PS       psCmd       `cmd:"" help:"List deployments."`
	Serve    serveCmd    `cmd:"" help:"Start the FTL server."`
	Call     callCmd     `cmd:"" help:"Call an FTL function."`
	Update   updateCmd   `cmd:"" help:"Update a deployment."`
	Kill     killCmd     `cmd:"" help:"Kill a deployment."`
	Schema   schemaCmd   `cmd:"" help:"FTL schema commands."`
	Deploy   deployCmd   `cmd:"" help:"Create a new deployment."`
	Download downloadCmd `cmd:"" help:"Download a deployment."`
}

var cli CLI

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.Configuration(kongtoml.Loader, ".ftl.toml", "~/.ftl.toml"),
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			node, ok := parent.(*kong.Command)
			if !ok {
				return nil
			}
			return &kong.Group{Key: node.Name, Title: "Command flags:"}
		}),
		kong.Vars{
			"version": version,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		},
	)

	rpc.InitialiseClients(cli.Authenticators)

	// Set the log level for child processes.
	os.Setenv("LOG_LEVEL", cli.LogConfig.Level.String())

	ctx, cancel := context.WithCancel(context.Background())

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Infof("FTL terminating with signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	kctx.Bind(cli.Endpoint)
	kctx.BindTo(ctx, (*context.Context)(nil))
	err := kctx.BindToProvider(makeDialer(ftlv1connect.NewVerbServiceClient))
	kctx.FatalIfErrorf(err)
	err = kctx.BindToProvider(makeDialer(ftlv1connect.NewControllerServiceClient))
	kctx.FatalIfErrorf(err)

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}

func makeDialer[Client rpc.Pingable](newClient func(connect.HTTPClient, string, ...connect.ClientOption) Client) func() (Client, error) {
	return func() (Client, error) {
		return rpc.Dial(newClient, cli.Endpoint.String(), log.Error), nil
	}
}
