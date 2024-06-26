// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package sql

import (
	"context"

	"github.com/alecthomas/types/optional"
)

const getModuleConfiguration = `-- name: GetModuleConfiguration :one
SELECT value
FROM module_configuration
WHERE
  (module IS NULL OR module = $1)
  AND name = $2
ORDER BY module NULLS LAST
LIMIT 1
`

func (q *Queries) GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error) {
	row := q.db.QueryRow(ctx, getModuleConfiguration, module, name)
	var value []byte
	err := row.Scan(&value)
	return value, err
}

const listModuleConfiguration = `-- name: ListModuleConfiguration :many
SELECT id, created_at, module, name, value
FROM module_configuration
ORDER BY module, name
`

func (q *Queries) ListModuleConfiguration(ctx context.Context) ([]ModuleConfiguration, error) {
	rows, err := q.db.Query(ctx, listModuleConfiguration)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ModuleConfiguration
	for rows.Next() {
		var i ModuleConfiguration
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.Module,
			&i.Name,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setModuleConfiguration = `-- name: SetModuleConfiguration :exec
INSERT INTO module_configuration (module, name, value)
VALUES ($1, $2, $3)
`

func (q *Queries) SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error {
	_, err := q.db.Exec(ctx, setModuleConfiguration, module, name, value)
	return err
}

const unsetModuleConfiguration = `-- name: UnsetModuleConfiguration :exec
DELETE FROM module_configuration
WHERE module = $1 AND name = $2
`

func (q *Queries) UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error {
	_, err := q.db.Exec(ctx, unsetModuleConfiguration, module, name)
	return err
}
