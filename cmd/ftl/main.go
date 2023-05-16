package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/bufbuild/connect-go"

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

	Devel    develCmd    `cmd:"" help:"Serve development FTL modules."`
	Serve    serveCmd    `cmd:"" help:"Start the FTL server."`
	Schema   schemaCmd   `cmd:"" help:"Retrieve the FTL schema."`
	List     listCmd     `cmd:"" help:"List all FTL functions."`
	Call     callCmd     `cmd:"" help:"Call an FTL function."`
	Go       goCmd       `cmd:"" help:"Commands specific to Go modules."`
	Deploy   deployCmd   `cmd:"" help:"Create a new deployment."`
	Download downloadCmd `cmd:"" help:"Download a deployment."`
	InitDB   initDBCmd   `cmd:"" name:"initdb" help:"Initialise the FTL database."`
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			node, ok := parent.(*kong.Command)
			if !ok {
				return nil
			}
			return &kong.Group{Key: node.Name, Title: "Command flags:"}
		}),
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
	err := kctx.BindToProvider(makeDialer(ftlv1connect.NewVerbServiceClient))
	kctx.FatalIfErrorf(err)
	err = kctx.BindToProvider(makeDialer(ftlv1connect.NewDevelServiceClient))
	kctx.FatalIfErrorf(err)
	err = kctx.BindToProvider(makeDialer(ftlv1connect.NewBackplaneServiceClient))
	kctx.FatalIfErrorf(err)

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}

func makeDialer[Client rpc.Pingable](newClient func(connect.HTTPClient, string, ...connect.ClientOption) Client) func() (Client, error) {
	return func() (Client, error) {
		return rpc.Dial(newClient, cli.Endpoint.URL()), nil
	}
}
