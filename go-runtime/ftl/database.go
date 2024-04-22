package ftl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"

	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver
)

type Database struct {
	Name string
}

// PostgresDatabase returns a handler for the named database.
func PostgresDatabase(name string) Database {
	return Database{
		Name: name,
	}
}

func (d Database) String() string { return fmt.Sprintf("database %q", d.Name) }

// Get returns the sql db connection for the database.
func (d Database) Get(ctx context.Context) *sql.DB {
	provider := modulecontext.DBProviderFromContext(ctx)
	db, err := provider.Get(d.Name)
	if err != nil {
		panic(err.Error())
	}
	return db
}
