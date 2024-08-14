package databasetesting

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/log"
)

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
		conn, err = stdsql.Open("pgx", noDBDSN.String())
		if err == nil {
			defer conn.Close()
			break
		}
		logger.Debugf("Waiting for database to be ready: %v", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

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
			return nil, err
		}

		_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
		if err != nil {
			return nil, err
		}
	}

	_, _ = conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName)) //nolint:errcheck // PG doesn't support "IF NOT EXISTS" so instead we just ignore any error.

	err = sql.Migrate(ctx, dsn, log.Debug)
	if err != nil {
		return nil, err
	}

	realConn, err := stdsql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	// Reset transient state in the database to a clean state for development purposes.
	// This includes things like resetting the state of async calls, leases,
	// controller/runner registration, etc. but not anything more.
	if !recreate {
		_, err = realConn.ExecContext(ctx, `
			WITH deleted AS (
				DELETE FROM async_calls
				RETURNING 1
			), deleted_fsm_instances AS (
				DELETE FROM fsm_instances
				RETURNING 1
			), deleted_leases AS (
				DELETE FROM leases
				RETURNING 1
			), deleted_controllers AS (
				DELETE FROM controller
				RETURNING 1
			), deleted_runners AS (
				DELETE FROM runners
				RETURNING 1
			)
			SELECT COUNT(*)
		`)
		if err != nil {
			return nil, err
		}
	}

	return realConn, nil
}
