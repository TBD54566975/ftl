package dev

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
)

const postgresContainerName = "ftl-db-1"
const mysqlContainerName = "ftl-mysql-1"

func SetupPostgres(ctx context.Context, image string, port int, recreate bool) (string, error) {
	dsn, err := SetupDatabase(ctx, image, port, postgresContainerName, 5432, WaitForPostgresReady, container.RunPostgres)
	if err != nil {
		return "", fmt.Errorf("failed to create database: %w", err)
	}
	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return "", fmt.Errorf("failed to create database: %w", err)
	}
	return dsn, nil
}

func SetupMySQL(ctx context.Context, image string, port int) (string, error) {
	return SetupDatabase(ctx, image, port, mysqlContainerName, 3306, WaitForMySQLReady, container.RunMySQL)
}

func SetupDatabase(ctx context.Context, image string, port int, containerName string, containerPort int, waitForReady func(ctx context.Context, port int) (string, error), runContainer func(ctx context.Context, name string, port int, image string) error) (string, error) {
	logger := log.FromContext(ctx)

	exists, err := container.DoesExist(ctx, containerName, optional.Some(image))
	if err != nil {
		return "", fmt.Errorf("failed to check if container exists: %w", err)
	}

	if !exists {
		logger.Debugf("Creating docker container '%s' for db", containerName)

		// check if port s.DBPort is already in use
		if l, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return "", fmt.Errorf("port %d is already in use", port)
		} else if err = l.Close(); err != nil {
			return "", fmt.Errorf("failed to close listener: %w", err)
		}

		err = runContainer(ctx, containerName, port, image)
		if err != nil {
			return "", fmt.Errorf("failed to run db container: %w", err)
		}

	} else {
		// Start the existing container
		err = container.Start(ctx, containerName)
		if err != nil {
			return "", fmt.Errorf("failed to start existing db container: %w", err)
		}

		// Grab the port from the existing container
		port, err = container.GetContainerPort(ctx, containerName, containerPort)
		if err != nil {
			return "", fmt.Errorf("failed to get port from existing db container: %w", err)
		}

		logger.Debugf("Reusing existing docker container %s on port %d for db", containerName, port)
	}

	dsn, err := waitForReady(ctx, port)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	return dsn, nil
}

func WaitForPostgresReady(ctx context.Context, port int) (string, error) {
	logger := log.FromContext(ctx)
	err := container.PollContainerHealth(ctx, postgresContainerName, 10*time.Minute)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	dsn := dsn.PostgresDSN("ftl", dsn.Port(port))
	logger.Debugf("Postgres DSN: %s", dsn)
	return dsn, nil
}

func WaitForMySQLReady(ctx context.Context, port int) (string, error) {
	logger := log.FromContext(ctx)
	err := container.PollContainerHealth(ctx, mysqlContainerName, 10*time.Minute)
	if err != nil {
		return "", fmt.Errorf("db container failed to be healthy: %w", err)
	}

	dsn := dsn.MySQLDSN("ftl", dsn.Port(port))
	logger.Debugf("MySQL DSN: %s", dsn)
	return dsn, nil
}
