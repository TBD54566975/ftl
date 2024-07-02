//go:build integration

package sql_test

import (
	"testing"

	in "github.com/TBD54566975/ftl/integration"
)

func TestDatabase(t *testing.T) {
	in.Run(t, "testdata/go/database/ftl-project.toml",
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
