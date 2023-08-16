package databasetesting

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
)

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
	workingDir := strings.TrimSpace(string(output))
	cmd := exec.Command(ctx, log.Debug, workingDir, "dbmate", "--url="+dsn, "--migrations-dir=backend/controller/internal/sql/schema", "up")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}
	return nil
}
