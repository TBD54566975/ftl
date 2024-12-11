package dev

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

//go:embed docker-compose.redpanda.yml
var redpandaDockerCompose string

func SetUpRedPanda(ctx context.Context) error {
	projCfg, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return fmt.Errorf("failed to get project config path")
	}
	err := container.ComposeUp(ctx, filepath.Dir(projCfg), "redpanda", redpandaDockerCompose)
	if err != nil {
		return fmt.Errorf("could not start redpanda: %w", err)
	}
	return nil
}
