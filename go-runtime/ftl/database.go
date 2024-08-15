package ftl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alecthomas/types/once"
	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver

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
			db, err := sql.Open("pgx", dsn)
			if err != nil {
				return nil, fmt.Errorf("failed to open database %q: %w", name, err)
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
