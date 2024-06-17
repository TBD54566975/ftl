package sql

type DBI interface {
	Querier
	Conn() ConnI
}

type ConnI interface {
	DBTX
}

type DB struct {
	conn ConnI
	*Queries
}

func NewDB(conn ConnI) *DB {
	return &DB{conn: conn, Queries: New(conn)}
}

func (d *DB) Conn() ConnI { return d.conn }
