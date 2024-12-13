package dev

import (
	"context"
	stdsql "database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/container"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
)

//go:embed docker-compose.mysql.yml
var mysqlDockerCompose string

//go:embed docker-compose.postgres.yml
var postgresDockerCompose string

// CreateForDevel creates and migrates a new database for development or testing.
//
// If "recreate" is true, the database will be dropped and recreated.
func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*stdsql.DB, error) {
	logger := log.FromContext(ctx)
	config, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	noDBDSN := *config
	noDBDSN.Path = "" // Remove the database name.

	var conn *stdsql.DB
	for range 10 {
		conn, err = observability.OpenDBAndInstrument(noDBDSN.String())
		if err == nil {
			defer conn.Close()
			break
		}
		logger.Debugf("Waiting for database to be ready: %v", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ctx canclled %w", ctx.Err())

		case <-time.After(1 * time.Second):
		}
	}
	if conn == nil {
		return nil, fmt.Errorf("database not ready after 10 tries: %w", err)
	}

	dbName := strings.TrimPrefix(config.Path, "/")

	if recreate {
		// Terminate any dangling connections.
		_, err = conn.ExecContext(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
			dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to terminate connections: %w", err)
		}

		_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
		if err != nil {
			return nil, fmt.Errorf("failed to drop database: %w", err)
		}
	}

	_, _ = conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName)) //nolint:errcheck // PG doesn't support "IF NOT EXISTS" so instead we just ignore any error.

	err = sql.Migrate(ctx, dsn, log.Debug)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	realConn, err := observability.OpenDBAndInstrument(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	realConn.SetMaxIdleConns(20)
	realConn.SetMaxOpenConns(20)
	// Reset transient state in the database to a clean state for development purposes.
	// This includes things like resetting the state of async calls, leases,
	// controller/runner registration, etc. but not anything more.
	if !recreate {
		_, err = realConn.ExecContext(ctx, `
			WITH deleted AS (
				DELETE FROM async_calls
				RETURNING 1
			)
			SELECT COUNT(*)
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to reset async_calls: %w", err)
		}
	}

	return realConn, nil
}

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
	err := container.ComposeUp(ctx, "postgres", postgresDockerCompose, envars...)
	if err != nil {
		return fmt.Errorf("could not start postgres: %w", err)
	}
	dsn := PostgresDSN(ctx, port)
	_, err = CreateForDevel(ctx, dsn, recreate)
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
	err := container.ComposeUp(ctx, "mysql", mysqlDockerCompose, envars...)
	if err != nil {
		return "", fmt.Errorf("could not start mysql: %w", err)
	}
	dsn := dsn.MySQLDSN("ftl", dsn.Port(port))
	log.FromContext(ctx).Debugf("MySQL DSN: %s", dsn)
	return dsn, nil
}
