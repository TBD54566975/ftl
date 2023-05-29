package sql

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateForDevel creates and migrates a new database for development or testing.
//
// If "recreate" is true, the database will be dropped and recreated.
func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*pgxpool.Pool, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	noDBDSN := config.Copy()
	noDBDSN.Database = ""
	conn, err := pgx.ConnectConfig(ctx, noDBDSN)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer conn.Close(ctx)

	if recreate {
		// Terminate any dangling connections.
		_, err = conn.Exec(ctx, `
		SELECT pid, pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()`,
			config.Database)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", config.Database))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	// PG doesn't support "IF NOT EXISTS" so instead we just ignore any error.
	_, _ = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", config.Database))

	err = Migrate(ctx, dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	realConn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return realConn, nil
}
