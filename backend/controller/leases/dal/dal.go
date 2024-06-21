// Package dal provides a data abstraction layer for leases
package dal

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backend/controller/leases/sql"
)

type DAL struct {
	db sql.DBI
}

func New(pool *pgxpool.Pool) *DAL {
	return &DAL{db: sql.NewDB(pool)}
}
