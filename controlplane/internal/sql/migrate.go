package sql

import (
	"context"
	"database/sql"
	"embed"

	"github.com/alecthomas/errors"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver
	"github.com/pressly/goose/v3"
)

//go:embed schema/*.sql
var migrations embed.FS

// Migrate a database connection to the latest schema using Goose.
func Migrate(ctx context.Context, dsn string) error {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer conn.Close()
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	err = goose.Up(conn, "schema")
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}
	return nil
}
