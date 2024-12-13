//go:build integration

package sql_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	in "github.com/TBD54566975/ftl/internal/integration"
)

func TestDatabase(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		// deploy real module against "testdb"
		in.CopyModule("database"),
		in.Deploy("database"),
		in.Call[in.Obj, in.Obj]("database", "insert", in.Obj{"data": "hello", "id": 1}, nil),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),

		// run tests which should only affect "testdb_test"
		in.IfLanguage("go", in.ExecModuleTest("database")),
		in.QueryRow("database_testdb", "SELECT data FROM requests", "hello"),
	)
}

func TestMySQL(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go", "java"),
		// deploy real module against "testdb"
		in.CopyModule("mysql"),
		in.Deploy("mysql"),
		in.Call[in.Obj, in.Obj]("mysql", "insert", in.Obj{"data": "hello"}, nil),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
		in.IfLanguage("go", in.ExecModuleTest("mysql")),
		in.Call[in.Obj, in.Obj]("mysql", "query", map[string]any{}, func(t testing.TB, response in.Obj) {
			assert.Equal(t, "hello", response["data"])
		}),
	)
}
