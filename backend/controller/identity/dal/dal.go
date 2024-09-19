package dal

import (
	"github.com/TBD54566975/ftl/backend/controller/identity/dal/internal/sql"
	"github.com/TBD54566975/ftl/backend/libdal"
)

type DAL struct {
	*libdal.Handle[DAL]
	db sql.Querier
}

func New(conn libdal.Connection) *DAL {
	return &DAL{
		db: sql.New(conn),
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{Handle: h, db: sql.New(h.Connection)}
		}),
	}
}
