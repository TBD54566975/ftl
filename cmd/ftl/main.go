package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/alecthomas/kong"
)

var version = "dev"

var cli struct {
	Version   kong.VersionFlag `help:"Show version information."`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`
}

func main() {
	kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.HelpOptions{
			Compact: true,
		},
		kong.Vars{
			"version": version,
		},
	)

	ctx := context.Background()

	// Handle SIGINT.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := log.New(cli.LogConfig, os.Stderr)
	ctx = log.ContextWithLogger(ctx, logger)
}
