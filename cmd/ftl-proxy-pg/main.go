package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

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
	t, err := strconv.ParseInt(ftl.Timestamp, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp %q: %s", ftl.Timestamp, err))
	}
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.Version, "timestamp": time.Unix(t, 0).Format(time.RFC3339)},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, false, "", "ftl-provisioner", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	proxy := pgproxy.New(cli.Config, func(ctx context.Context, params map[string]string) (string, error) {
		return "postgres://localhost:5432/postgres?user=" + params["user"], nil
	})
	if err := proxy.Start(ctx); err != nil {
		kctx.FatalIfErrorf(err, "failed to start proxy")
	}
}
