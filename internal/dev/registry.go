package dev

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/log"
)

const ftlRegistryName = "ftl-registry-1"

func SetupRegistry(ctx context.Context, image string, port int) error {
	logger := log.FromContext(ctx)

	exists, err := container.DoesExist(ctx, ftlRegistryName, optional.Some(image))
	if err != nil {
		return fmt.Errorf("failed to check if container exists: %w", err)
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for registry", ftlRegistryName)

		// check if port is already in use
		if l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return fmt.Errorf("port %d is already in use", port)
		} else if err = l.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}

		err = container.Run(ctx, image, ftlRegistryName, port, 5000, optional.None[string]())
		if err != nil {
			return fmt.Errorf("failed to run registry container: %w", err)
		}

	} else {
		// Start the existing container
		err = container.Start(ctx, ftlRegistryName)
		if err != nil {
			return fmt.Errorf("failed to start existing registry container: %w", err)
		}

		// Grab the port from the existing container
		port, err = container.GetContainerPort(ctx, ftlRegistryName, 5000)
		if err != nil {
			return fmt.Errorf("failed to get port from existing registry container: %w", err)
		}

		logger.Debugf("Reusing existing docker container %s on port %d for image registry", ftlRegistryName, port)
	}

	err = WaitForRegistryReady(ctx, port)
	if err != nil {
		return fmt.Errorf("registry container failed to be healthy: %w", err)
	}

	return nil
}

func WaitForRegistryReady(ctx context.Context, port int) error {

	timeout := time.After(10 * time.Minute)
	retry := time.NewTicker(5 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled waiting for registry")
		case <-timeout:
			return fmt.Errorf("timed out waiting for registry to be healthy")
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
