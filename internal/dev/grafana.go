package dev

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

//go:embed docker-compose.grafana.yml
var grafanaDockerCompose string

func SetupGrafana(ctx context.Context, image string) error {
	projCfg, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return fmt.Errorf("failed to get project config path")
	}
	err := container.ComposeUp(ctx, filepath.Dir(projCfg), "grafana", grafanaDockerCompose)
	if err != nil {
		return fmt.Errorf("could not start grafana: %w", err)
	}
	err = WaitForPortReady(ctx, 3000)
	if err != nil {
		return fmt.Errorf("registry container failed to be healthy: %w", err)
	}
	return nil
}
