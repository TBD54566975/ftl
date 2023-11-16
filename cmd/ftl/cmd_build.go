package main

import (
	"context"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/alecthomas/errors"
)

type buildCmd struct {
	Base string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context) error {
	// Load the TOML file.
	config, err := moduleconfig.LoadConfig(b.Base)
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
		buildCmd = "source ../bin/activate-hermit && mvn compile"
	}

	logger.Infof("Building kotlin module '%s'", config.Module)
	logger.Infof("Using build command '%s'", buildCmd)

	// Have to activate hermit within the same shell
	err := exec.Command(ctx, log.Debug, b.Base, "bash", "-c", buildCmd).Run()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
