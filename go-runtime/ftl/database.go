package ftl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/alecthomas/types/once"
	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/TBD54566975/ftl/internal/modulecontext"
)

type Database struct {
	Name   string
	DBType modulecontext.DBType

	db *once.Handle[*sql.DB]
}

// PostgresDatabase returns a handler for the named database.
func PostgresDatabase(name string) Database {
	return Database{
		Name:   name,
		DBType: modulecontext.DBTypePostgres,
		db: once.Once(func(ctx context.Context) (*sql.DB, error) {
			provider := modulecontext.FromContext(ctx).CurrentContext()
			dsn, err := provider.GetDatabase(name, modulecontext.DBTypePostgres)
			if err != nil {
				return nil, fmt.Errorf("failed to get database %q: %w", name, err)
			}
			db, err := otelsql.Open("pgx", dsn)
			if err != nil {
				return nil, fmt.Errorf("failed to open database %q: %w", name, err)
			}

			// sets db.system and db.name attributes
			metricAttrs := otelsql.WithAttributes(
				semconv.DBSystemPostgreSQL,
				semconv.DBNameKey.String(name),
				attribute.Bool("ftl.is_user_service", true),
				attribute.String("ftl.sql.module.name", provider.module),
			)
			err = otelsql.RegisterDBStatsMetrics(db, metricAttrs)
			if err != nil {
				return nil, fmt.Errorf("failed to register database metrics: %w", err)
			}
			return db, nil
		}),
	}
}

func (d Database) String() string { return fmt.Sprintf("database %q", d.Name) }

// Get returns the SQL DB connection for the database.
func (d Database) Get(ctx context.Context) *sql.DB {
	db, err := d.db.Get(ctx)
	if err != nil {
		panic(err)
	}
	return db
}
