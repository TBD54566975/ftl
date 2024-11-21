package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/runner"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/pgproxy"
)

var cli struct {
	Version      kong.VersionFlag `help:"Show version."`
	LogConfig    log.Config       `prefix:"log-" embed:""`
	RunnerConfig runner.Config    `embed:""`
	ProxyConfig  pgproxy.Config   `embed:"" prefix:"pgproxy-"`
}

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	kctx := kong.Parse(&cli, kong.Description(`
FTL - Towards a 𝝺-calculus for large-scale systems

The Runner is the component of FTL that coordinates with the Controller to spawn
and route to user code.
	`), kong.Vars{
		"version":       ftl.Version,
		"deploymentdir": filepath.Join(cacheDir, "ftl-runner", "${runner}", "deployments"),
	})
	// Substitute in the runner key into the deployment directory.
	cli.RunnerConfig.DeploymentDir = os.Expand(cli.RunnerConfig.DeploymentDir, func(key string) string {
		if key == "runner" {
			return cli.RunnerConfig.Key.String()
		}
		return key
	})
	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return runPGProxy(ctx, cli.ProxyConfig)
	})
	g.Go(func() error {
		return runner.Start(ctx, cli.RunnerConfig)
	})
	kctx.FatalIfErrorf(g.Wait())
}

func runPGProxy(ctx context.Context, config pgproxy.Config) error {
	if err := pgproxy.New(config, func(ctx context.Context, params map[string]string) (string, error) {
		return "postgres://127.0.0.1:5432/postgres?user=" + params["user"], nil
	}).Start(ctx); err != nil {
		return fmt.Errorf("failed to start pgproxy: %w", err)
	}
	return nil
}
