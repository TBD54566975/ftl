package dev

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/TBD54566975/ftl/internal/container"
)

//go:embed docker-compose.redpanda.yml
var redpandaDockerCompose string

func SetUpRedPanda(ctx context.Context) error {
	err := container.ComposeUp(ctx, "redpanda", redpandaDockerCompose)
	if err != nil {
		return fmt.Errorf("could not start redpanda: %w", err)
	}
	return nil
}
