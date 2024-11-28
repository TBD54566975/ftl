package dev

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "embed"

	"github.com/TBD54566975/ftl/internal/container"
)

//go:embed docker-compose.registry.yml
var registryDockerCompose string

func SetupRegistry(ctx context.Context, image string, port int) error {
	err := container.ComposeUp(ctx, "registry", registryDockerCompose,
		"FTL_REGISTRY_IMAGE="+image,
		"FTL_REGISTRY_PORT="+strconv.Itoa(port))
	if err != nil {
		return fmt.Errorf("could not start registry: %w", err)
	}
	err = WaitForPortReady(ctx, port)
	if err != nil {
		return fmt.Errorf("registry container failed to be healthy: %w", err)
	}
	return nil
}

func WaitForPortReady(ctx context.Context, port int) error {
	timeout := time.After(10 * time.Minute)
	retry := time.NewTicker(5 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled waiting for container")
		case <-timeout:
			return fmt.Errorf("timed out waiting for container to be healthy")
		case <-retry.C:
			url := fmt.Sprintf("http://127.0.0.1:%d", port)

			req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil) //nolint:gosec
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil

			}
		}

	}
}
