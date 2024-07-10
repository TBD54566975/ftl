//go:build integration

package sql_test

import (
	"fmt"
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestDatabase(t *testing.T) {
	in.Run(t, "database/ftl-project.toml",
		// deploy real module against "testdb"
		in.CopyModule("database"),
		in.CreateDBAction("database", "testdb", false),
		in.Deploy("database"),
		in.Call("database", "insert", in.Obj{"data": "hello"}, nil),
		in.QueryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		in.CreateDBAction("database", "testdb", true),
		in.ExecModuleTest("database"),
		in.QueryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestMigrate(t *testing.T) {
	dbName := "ftl_test"
	dbUri := fmt.Sprintf("postgres://postgres:secret@localhost:15432/%s?sslmode=disable", dbName)

	q := func() in.Action {
		return in.QueryRow(dbName, "SELECT version FROM schema_migrations WHERE version = '20240704103403'", "20240704103403")
	}

	in.RunWithoutController(t, "",
		in.DropDBAction(t, dbName),
		in.Fail(q(), "Should fail because the database does not exist."),
		in.Exec("ftl", "migrate", "--dsn", dbUri),
		q(),
	)
}
