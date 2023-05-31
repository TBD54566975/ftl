package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	_ "github.com/TBD54566975/ftl/common/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/runner"
)

var version = "dev"

var config struct {
	Version      kong.VersionFlag `help:"Show version."`
	LogConfig    log.Config       `prefix:"log-" embed:""`
	RunnerConfig runner.Config    `embed:""`
}

func main() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	kctx := kong.Parse(&config, kong.Description(`
FTL - Towards a 𝝺-calculus for large-scale systems

The Runner is the component of FTL that coordinates with the ControlPlane to spawn
and route to user code.
	`), kong.Vars{
		"version":       version,
		"deploymentdir": filepath.Join(cacheDir, "ftl-runner-go", "deployments"),
	})
	logger := log.Configure(os.Stderr, config.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	err = runner.Start(ctx, config.RunnerConfig)
	kctx.FatalIfErrorf(err)
}
