package main

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type buildCmd struct {
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (b *buildCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	// Load the TOML file.
	config, err := moduleconfig.LoadConfig(b.ModuleDir)
	if err != nil {
		return err
	}

	switch config.Language {
	case "kotlin":
		return b.buildKotlin(ctx, config)

	case "go":
		return b.buildGo(ctx, config, client)

	default:
		return fmt.Errorf("unable to build, unknown language %q", config.Language)
	}
}
