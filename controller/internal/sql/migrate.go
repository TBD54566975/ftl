package sql

import (
	"context"
	"database/sql"
	"embed"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/internal/exec"
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
	output, err := exec.Capture(ctx, ".", "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return errors.Wrap(err, "failed to find git root")
	}
	cmd := exec.Command(ctx, strings.TrimSpace(string(output)), "dbmate", "--url="+dsn, "--migrations-dir=controller/internal/sql/schema", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}
	return nil
}
