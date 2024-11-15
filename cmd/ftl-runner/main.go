package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/runner"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
)

var cli struct {
	Version      kong.VersionFlag `help:"Show version."`
	LogConfig    log.Config       `prefix:"log-" embed:""`
	RunnerConfig runner.Config    `embed:""`
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
	err = runner.Start(ctx, cli.RunnerConfig)
	kctx.FatalIfErrorf(err)
}
