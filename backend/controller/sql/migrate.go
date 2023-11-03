package sql

import (
	"context"
	"database/sql"
	"embed"
	"net/url"

	"github.com/alecthomas/errors"
	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/backend/common/log"
)

//go:embed schema
var schema embed.FS

// Migrate the database.
func Migrate(ctx context.Context, dsn string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return errors.Wrap(err, "invalid DSN")
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer conn.Close()

	db := dbmate.New(u)
	db.FS = schema
	db.Log = log.FromContext(ctx).Scope("migrate").WriterAt(log.Debug)
	db.MigrationsDir = []string{"schema"}
	err = db.CreateAndMigrate()
	if err != nil {
		return errors.Wrap(err, "failed to create and migrate database")
	}
	return nil
}
