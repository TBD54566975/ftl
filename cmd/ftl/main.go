package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/common/log"
)

var version = "dev"

var cli struct {
	Version   kong.VersionFlag `help:"Show version information."`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`

	Serve serveCmd `cmd:"" help:"Serve a directory of FTL functions."`
	// List  listCmd  `cmd:"" help:"List all FTL functions."`
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

	kctx.BindTo(ctx, (*context.Context)(nil))

	err := kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}
