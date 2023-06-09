package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/observability"
	"github.com/TBD54566975/ftl/runner"
)

var version = "dev"

var config struct {
	Version             kong.VersionFlag     `help:"Show version."`
	LogConfig           log.Config           `prefix:"log-" embed:""`
	ObservabilityConfig observability.Config `embed:"" prefix:"observability-"`
	RunnerConfig        runner.Config        `embed:""`
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
		"deploymentdir": filepath.Join(cacheDir, "ftl-runner", "deployments"),
	})
	logger := log.Configure(os.Stderr, config.LogConfig)
	ctx := log.ContextWithLogger(context.Background(), logger)
	if config.ObservabilityConfig.Endpoint == nil {
		config.ObservabilityConfig.Endpoint = config.RunnerConfig.FTLEndpoint
	}
	err = observability.Init(ctx, "ftl-runner", config.ObservabilityConfig)
	kctx.FatalIfErrorf(err)
	err = runner.Start(ctx, config.RunnerConfig)
	kctx.FatalIfErrorf(err)
}
