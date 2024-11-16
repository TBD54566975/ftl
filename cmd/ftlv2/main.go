package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
)

var cli struct {
	LogConfig log.Config `embed:"" prefix:"log-" help:"Configure logging."`
	Dev       devCmd     `cmd:"" help:"Run the FTL development server."`
}

func main() {
	ctx := context.Background()

	kctx := kong.Parse(&cli, kong.Vars{
		"version": ftl.Version,
	},
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
	)
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	// Bind values.
	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(logger)

	err := kctx.Run()
	kctx.FatalIfErrorf(err)
}
