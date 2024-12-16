package dev

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/block/ftl/internal/container"
)

//go:embed docker-compose.grafana.yml
var grafanaDockerCompose string

func SetupGrafana(ctx context.Context, image string) error {
	err := container.ComposeUp(ctx, "grafana", grafanaDockerCompose)
	if err != nil {
		return fmt.Errorf("could not start grafana: %w", err)
	}
	err = WaitForPortReady(ctx, 3000)
	if err != nil {
		return fmt.Errorf("registry container failed to be healthy: %w", err)
	}
	return nil
}
