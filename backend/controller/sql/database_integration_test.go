//go:build integration

package sql_test

import (
	"fmt"
	"testing"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestDatabase(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		in.WithFTLConfig("database/ftl-project.toml"),
		// deploy real module against "testdb"
		in.CopyModule("database"),
		in.CreateDBAction("database", "testdb", false),
		in.Deploy("database"),
		in.Call[in.Obj, in.Obj]("database", "insert", in.Obj{"data": "hello"}, nil),
		in.QueryRow("testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		in.CreateDBAction("database", "testdb", true),
		in.IfLanguage("go", in.ExecModuleTest("database")),
		in.QueryRow("testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestMigrate(t *testing.T) {
	dbName := "ftl_test"
	dbUri := fmt.Sprintf("postgres://postgres:secret@localhost:15432/%s?sslmode=disable", dbName)

	q := func() in.Action {
		return in.QueryRow(dbName, "SELECT version FROM schema_migrations WHERE version = '20240704103403'", "20240704103403")
	}

	in.Run(t,
		in.WithoutController(),
		in.DropDBAction(t, dbName),
		in.Fail(q(), "Should fail because the database does not exist."),
		in.Exec("ftl", "migrate", "--dsn", dbUri),
		q(),
	)
}
