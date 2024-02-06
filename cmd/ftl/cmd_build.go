package main

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type buildCmd struct {
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)

	startTime := time.Now()
	// Load the TOML file.
	var err error
	config, err := moduleconfig.LoadConfig(b.ModuleDir)
	if err != nil {
		return err
	}

	logger.Infof("Building %s module '%s'", config.Language, config.Module)

	switch config.Language {
	case "kotlin":
		err = b.buildKotlin(ctx, config)

	case "go":
		err = b.buildGo(ctx, client)

	default:
		return fmt.Errorf("unable to build, unknown language %q", config.Language)
	}

	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	formattedDuration := fmt.Sprintf("%.2fs", duration.Seconds())
	logger.Infof("Successfully built module '%s' in %s", config.Module, formattedDuration)
	return nil
}
