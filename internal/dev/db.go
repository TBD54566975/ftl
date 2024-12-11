package dev

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

//go:embed docker-compose.mysql.yml
var mysqlDockerCompose string

//go:embed docker-compose.postgres.yml
var postgresDockerCompose string

func PostgresDSN(ctx context.Context, port int) string {
	if port == 0 {
		port = 15432
	}
	return dsn.PostgresDSN("ftl", dsn.Port(port))
}

func SetupPostgres(ctx context.Context, image optional.Option[string], port int, recreate bool) error {
	envars := []string{}
	if port != 0 {
		envars = append(envars, "POSTGRES_PORT="+strconv.Itoa(port))
	}
	if imaneName, ok := image.Get(); ok {
		envars = append(envars, "FTL_DATABASE_IMAGE="+imaneName)
	}
	projCfg, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return fmt.Errorf("failed to get project config path")
	}
	err := container.ComposeUp(ctx, filepath.Dir(projCfg), "postgres", postgresDockerCompose, envars...)
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
	envars := []string{}
	if port != 0 {
		envars = append(envars, "MYSQL_PORT="+strconv.Itoa(port))
	}
	projCfg, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return "", fmt.Errorf("failed to get project config path")
	}
	err := container.ComposeUp(ctx, filepath.Dir(projCfg), "mysql", mysqlDockerCompose, envars...)
	if err != nil {
		return "", fmt.Errorf("could not start mysql: %w", err)
	}
	dsn := dsn.MySQLDSN("ftl", dsn.Port(port))
	log.FromContext(ctx).Debugf("MySQL DSN: %s", dsn)
	return dsn, nil
}
