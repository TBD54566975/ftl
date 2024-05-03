package ftl

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
)

type Database struct {
	Name   string
	DBType modulecontext.DBType
}

// PostgresDatabase returns a handler for the named database.
func PostgresDatabase(name string) Database {
	return Database{
		Name:   name,
		DBType: modulecontext.DBTypePostgres,
	}
}

func (d Database) String() string { return fmt.Sprintf("database %q", d.Name) }

// Get returns the sql db connection for the database.
func (d Database) Get(ctx context.Context) *sql.DB {
	provider := modulecontext.FromContext(ctx)
	db, err := provider.GetDatabase(d.Name, d.DBType)
	if err != nil {
		panic(err.Error())
	}
	return db
}
