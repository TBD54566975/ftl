package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/pgproxy"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`

	pgproxy.Config
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.FormattedVersion},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err := observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	proxy := pgproxy.New(cli.Config.Listen, func(ctx context.Context, params map[string]string) (string, error) {
		return "postgres://localhost:5432/postgres?user=" + params["user"], nil
	})
	if err := proxy.Start(ctx, nil); err != nil {
		kctx.FatalIfErrorf(err, "failed to start proxy")
	}
}
