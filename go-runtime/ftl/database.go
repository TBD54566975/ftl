package ftl

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver
)

// PostgresDatabase returns a Postgres database connection for the named database.
func PostgresDatabase(name string) *sql.DB {
	module := strings.ToUpper(callerModule())
	key := fmt.Sprintf("FTL_POSTGRES_DSN_%s_%s", module, strings.ToUpper(name))
	dsn, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("missing DSN environment variable %s", key))
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %s", err))
	}
	return db
}
