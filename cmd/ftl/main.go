package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/socket"
)

var version = "dev"

var cli struct {
	Version   kong.VersionFlag `help:"Show version information."`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Socket    socket.Socket    `default:"tcp://127.0.0.1:8892" help:"FTL socket." env:"FTL_SOCKET"`

	Serve serveCmd `cmd:"" help:"Serve FTL modules."`
	List  listCmd  `cmd:"" help:"List all FTL functions."`
	Call  callCmd  `cmd:"" help:"Call an FTL function."`
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

	logger := log.New(cli.LogConfig, os.Stderr).With("C", "FTL")
	ctx = log.ContextWithLogger(ctx, logger)

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Info("FTL terminating", "signal", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	kctx.Bind(cli.Socket)
	kctx.BindTo(ctx, (*context.Context)(nil))
	err := kctx.BindToProvider(dialAgent(ctx))
	kctx.FatalIfErrorf(err)

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}

func dialAgent(ctx context.Context) func() (ftlv1.VerbServiceClient, error) {
	return func() (ftlv1.VerbServiceClient, error) {
		conn, err := grpc.DialContext(ctx, cli.Socket.String(),
			// grpc.WithBlock(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(socket.Dialer))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return ftlv1.NewVerbServiceClient(conn), nil
	}
}
