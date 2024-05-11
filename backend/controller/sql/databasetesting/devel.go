package databasetesting

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/internal/log"
)

// CreateForDevel creates and migrates a new database for development or testing.
//
// If "recreate" is true, the database will be dropped and recreated.
func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*pgxpool.Pool, error) {
	logger := log.FromContext(ctx)
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	noDBDSN := config.Copy()
	noDBDSN.Database = ""
	var conn *pgx.Conn
	for range 10 {
		conn, err = pgx.ConnectConfig(ctx, noDBDSN)
		if err == nil {
			defer conn.Close(ctx)
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

	if recreate {
		// Terminate any dangling connections.
		_, err = conn.Exec(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
			config.Database)
		if err != nil {
			return nil, err
		}

		_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", config.Database))
		if err != nil {
			return nil, err
		}
	}

	// PG doesn't support "IF NOT EXISTS" so instead we just ignore any error.
	_, _ = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", config.Database))

	err = sql.Migrate(ctx, dsn)
	if err != nil {
		return nil, err
	}

	realConn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	// Reset transient state in the database to a clean state for development purposes.
	// This includes things like resetting the state of async calls, leases,
	// controller/runner registration, etc. but not anything more.
	if !recreate {
		_, err = realConn.Exec(ctx, `
			WITH deleted AS (
				DELETE FROM async_calls
				RETURNING 1
			), deleted_fsm_executions AS (
				DELETE FROM fsm_executions
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
