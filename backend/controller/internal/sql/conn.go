package sql

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/jackc/pgx/v5"
)

type DBI interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DB struct {
	conn DBI
	*Queries
}

func NewDB(conn DBI) *DB {
	return &DB{conn: conn, Queries: New(conn)}
}

func (d *DB) Conn() DBI { return d.conn }

func (d *DB) Begin(ctx context.Context) (*Tx, error) {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Tx{tx: tx, Queries: New(tx)}, nil
}

type Tx struct {
	tx pgx.Tx
	*Queries
}

func (t *Tx) Commit(ctx context.Context) error {
	return errors.WithStack(t.tx.Commit(ctx))
}

func (t *Tx) Rollback(ctx context.Context) error {
	return errors.WithStack(t.tx.Rollback(ctx))
}

// CommitOrRollback can be used in a defer statement to commit or rollback a
// transaction depending on whether the enclosing function returned an error.
//
//	func myFunc() (err error) {
//	  tx, err := db.Begin(ctx)
//	  if err != nil { return err }
//	  defer tx.CommitOrRollback(ctx, &err)
//	  ...
//	}
func (t *Tx) CommitOrRollback(ctx context.Context, err *error) {
	if *err != nil {
		_ = t.Rollback(ctx)
	} else {
		*err = errors.WithStack(t.Commit(ctx))
	}
}
