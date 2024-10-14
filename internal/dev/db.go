package dev

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dsn"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/log"
)

const ftlContainerName = "ftl-db-1"

func SetupDB(ctx context.Context, image string, port int, recreate bool) (string, error) {
	logger := log.FromContext(ctx)

	exists, err := container.DoesExist(ctx, ftlContainerName, optional.Some(image))
	if err != nil {
		return "", fmt.Errorf("failed to check if container exists: %w", err)
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for postgres db", ftlContainerName)

		// check if port s.DBPort is already in use
		if l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return "", fmt.Errorf("port %d is already in use", port)
		} else if err = l.Close(); err != nil {
			return "", fmt.Errorf("failed to close listener: %w", err)
		}

		err = container.RunDB(ctx, ftlContainerName, port, image)
		if err != nil {
			return "", fmt.Errorf("failed to run db container: %w", err)
		}

		recreate = true
	} else {
		// Start the existing container
		err = container.Start(ctx, ftlContainerName)
		if err != nil {
			return "", fmt.Errorf("failed to start existing db container: %w", err)
		}

		// Grab the port from the existing container
		port, err = container.GetContainerPort(ctx, ftlContainerName, 5432)
		if err != nil {
			return "", fmt.Errorf("failed to get port from existing db container: %w", err)
		}

		logger.Debugf("Reusing existing docker container %s on port %d for postgres db", ftlContainerName, port)
	}

	dsn, err := WaitForDBReady(ctx, port)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return "", fmt.Errorf("failed to create database: %w", err)
	}

	return dsn, nil
}

func WaitForDBReady(ctx context.Context, port int) (string, error) {
	logger := log.FromContext(ctx)
	err := container.PollContainerHealth(ctx, ftlContainerName, 10*time.Second)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	dsn := dsn.DSN("ftl", dsn.Port(port))
	logger.Debugf("Postgres DSN: %s", dsn)
	return dsn, nil
}
