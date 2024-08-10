package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestMigrate(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	db := sqltest.OpenForTesting(ctx, t)

	mfs, err := fs.Sub(migrations, "testdata")
	assert.NoError(t, err)

	err = Migrate(ctx, db, mfs, Migrate30280103221530SplitNameAge)
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

//go:embed testdata
var migrations embed.FS

func Migrate30280103221530SplitNameAge(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		ALTER TABLE test
			ADD COLUMN name TEXT,
			ADD COLUMN age INT
	`)
	if err != nil {
		return fmt.Errorf("failed to add columns: %w", err)
	}
	rows, err := tx.QueryContext(ctx, "SELECT id, name_and_age FROM test")
	if err != nil {
		return fmt.Errorf("failed to query test: %w", err)
	}
	defer rows.Close()
	type userUpdate struct {
		name string
		age  int64
	}
	updates := map[int]userUpdate{}
	for rows.Next() {
		var id int
		var name string
		var age int64
		err = rows.Scan(&id, &name)
		if err != nil {
			return fmt.Errorf("failed to scan user: %w", err)
		}
		nameAge := strings.Fields(name)
		name = nameAge[0]
		switch len(nameAge) {
		case 1:
		case 2:
			age, err = strconv.ParseInt(nameAge[1], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse age: %w", err)
			}
		default:
			return fmt.Errorf("invalid name %q", name)
		}
		// We can't update the table while iterating over it, so we store the updates in a map.
		updates[id] = userUpdate{name, age}
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("failed to close rows: %w", err)
	}
	for id, update := range updates {
		_, err = tx.ExecContext(ctx, "UPDATE test SET name = $1, age = $2 WHERE id = $3", update.name, update.age, id)
		if err != nil {
			return fmt.Errorf("failed to update user %d: %w", id, err)
		}
	}
	_, err = tx.ExecContext(ctx, `
		ALTER TABLE test
			DROP COLUMN name_and_age,
			ALTER COLUMN name SET NOT NULL,
			ALTER COLUMN age SET NOT NULL
  `)
	if err != nil {
		return fmt.Errorf("failed to drop column: %w", err)
	}
	return nil
}
