package migrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

func MigrateSplitNameAge(ctx context.Context, tx *sql.Tx) error {
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
	return nil
}
