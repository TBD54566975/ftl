package migrate

import (
	"context"
	"embed"
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/sql/migrate/migrationtest"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestMigrate(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	db := sqltest.OpenForTesting(ctx, t)

	mfs, err := fs.Sub(migrations, "migrationtest")
	assert.NoError(t, err)

	err = Migrate(ctx, db, mfs, Migration("30280103000000", "split_name_age", migrationtest.MigrateSplitNameAge))
	assert.NoError(t, err)

	rows, err := db.QueryContext(ctx, "SELECT name, age FROM test")
	assert.NoError(t, err)
	defer rows.Close()
	type user struct {
		name string
		age  int
	}
	actual := []user{}
	for rows.Next() {
		var u user
		assert.NoError(t, rows.Scan(&u.name, &u.age))
		actual = append(actual, u)
	}
	expected := []user{
		{"Alice", 30},
	}
	assert.Equal(t, expected, actual)
}

//go:embed migrationtest
var migrations embed.FS
