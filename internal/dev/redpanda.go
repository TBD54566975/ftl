package dev

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/container"
)

//go:embed docker-compose.redpanda.yml
var redpandaDockerCompose string

func SetUpRedPanda(ctx context.Context) error {
	var profile optional.Option[string]
	if _, ci := os.LookupEnv("CI"); !ci {
		// include console except in CI
		profile = optional.Some[string]("console")
	}
	err := container.ComposeUp(ctx, "redpanda", redpandaDockerCompose, profile)
	if err != nil {
		return fmt.Errorf("could not start redpanda: %w", err)
	}
	return nil
}
