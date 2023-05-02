package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/socket"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var version = "dev"

var cli struct {
	Version   kong.VersionFlag `help:"Show version information."`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Endpoint  socket.Socket    `default:"tcp://127.0.0.1:8892" help:"FTL endpoint to bind/connect to." env:"FTL_ENDPOINT"`

	Serve  serveCmd  `cmd:"" help:"Serve FTL modules."`
	Schema schemaCmd `cmd:"" help:"Retrieve the FTL schema."`
	List   listCmd   `cmd:"" help:"List all FTL functions."`
	Call   callCmd   `cmd:"" help:"Call an FTL function."`
	Go     goCmd     `cmd:"" help:"Commands specific to Go modules."`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{
			"version": version,
		},
	)

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
	err := kctx.BindToProvider(dialVerbService(ctx))
	kctx.FatalIfErrorf(err)
	err = kctx.BindToProvider(dialDevelService(ctx))
	kctx.FatalIfErrorf(err)

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}

func dialVerbService(ctx context.Context) func() (ftlv1connect.VerbServiceClient, error) {
	return func() (ftlv1connect.VerbServiceClient, error) {
		return rpc.Dial(ftlv1connect.NewVerbServiceClient, cli.Endpoint.URL()), nil
	}
}

func dialDevelService(ctx context.Context) func() (ftlv1connect.DevelServiceClient, error) {
	return func() (ftlv1connect.DevelServiceClient, error) {
		return rpc.Dial(ftlv1connect.NewDevelServiceClient, cli.Endpoint.URL()), nil
	}
}
