package main

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
)

type buildCmd struct {
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context) error {
	// Load the TOML file.
	config, err := moduleconfig.LoadConfig(b.ModuleDir)
	if err != nil {
		return errors.WithStack(err)
	}

	switch config.Language {
	case "kotlin":
		return b.buildKotlin(ctx, config)
	default:
		return errors.Errorf("unable to build. unknown language %q", config.Language)
	}
}

func (b *buildCmd) buildKotlin(ctx context.Context, config moduleconfig.ModuleConfig) error {
	logger := log.FromContext(ctx)

	buildCmd := config.Build

	if buildCmd == "" {
		buildCmd = "mvn compile"
	}

	logger.Infof("Building kotlin module '%s'", config.Module)
	logger.Infof("Using build command '%s'", buildCmd)

	err := exec.Command(ctx, logger.GetLevel(), b.ModuleDir, "bash", "-c", buildCmd).Run()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
