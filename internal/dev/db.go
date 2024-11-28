package dev

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
)

//go:embed docker-compose.mysql.yml
var mysqlDockerCompose string

//go:embed docker-compose.postgres.yml
var postgresDockerCompose string

func PostgresDSN(ctx context.Context, port int) string {
	return dsn.PostgresDSN("ftl", dsn.Port(port))
}

func SetupPostgres(ctx context.Context, image optional.Option[string], port int, recreate bool) error {
	envars := []string{"POSTGRES_PORT=" + strconv.Itoa(port)}
	if imaneName, ok := image.Get(); ok {
		envars = append(envars, "POSTGRES_IMAGE="+imaneName)
	}
	err := container.ComposeUp(ctx, "postgres", postgresDockerCompose, envars...)
	if err != nil {
		return fmt.Errorf("could not start postgres: %w", err)
	}
	dsn := PostgresDSN(ctx, port)
	_, err = databasetesting.CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	return nil
}

func SetupMySQL(ctx context.Context, port int) (string, error) {
	err := container.ComposeUp(ctx, "mysql", mysqlDockerCompose, "MYSQL_PORT="+strconv.Itoa(port))
	if err != nil {
		return "", fmt.Errorf("could not start mysql: %w", err)
	}
	dsn := dsn.MySQLDSN("ftl", dsn.Port(port))
	log.FromContext(ctx).Debugf("MySQL DSN: %s", dsn)
	return dsn, nil
}
