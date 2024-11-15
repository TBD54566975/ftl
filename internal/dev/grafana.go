package dev

import (
	"context"
	"fmt"
	"net"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/log"
)

const ftlGrafanaName = "ftl-otel-lgtm-1"

func SetupGrafana(ctx context.Context, image string) error {
	logger := log.FromContext(ctx)

	exists, err := container.DoesExist(ctx, ftlGrafanaName, optional.Some(image))
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for grafana", ftlGrafanaName)
		// check if port is already in use
		ports := []int{3000, 4317, 4318}
		for _, port := range ports {
			if l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
				return fmt.Errorf("port %d is already in use", port)
			} else if err = l.Close(); err != nil {
				return fmt.Errorf("failed to close listener: %w", err)
			}
		}
		err = container.Run(ctx, image, ftlGrafanaName, map[int]int{3000: 3000, 4317: 4317, 4318: 4318}, optional.None[string](), "ENABLE_LOGS_ALL=true", "GF_PATHS_DATA=/data/grafana")
		if err != nil {
			return fmt.Errorf("failed to run grafana container: %w", err)
		}

	} else {
		// Start the existing container
		err = container.Start(ctx, ftlGrafanaName)
		if err != nil {
			return fmt.Errorf("failed to start existing registry container: %w", err)
		}

		logger.Debugf("Reusing existing docker container %s for grafana", ftlGrafanaName)
	}

	err = WaitForPortReady(ctx, 3000)
	if err != nil {
		return fmt.Errorf("registry container failed to be healthy: %w", err)
	}

	return nil
}
