package sql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type DBI interface {
	Querier
	Conn() ConnI
	Begin(ctx context.Context) (*Tx, error)
}

type ConnI interface {
	DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

type DB struct {
	conn ConnI
	*Queries
}

func NewDB(conn ConnI) *DB {
	return &DB{conn: conn, Queries: New(conn)}
}

func (d *DB) Conn() ConnI { return d.conn }

func (d *DB) Begin(ctx context.Context) (*Tx, error) {
	tx, err := d.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, Queries: New(tx)}, nil
}

type Tx struct {
	tx         pgx.Tx
	savepoints []string
	*Queries
}

func (t *Tx) Conn() ConnI { return t.tx }

func (t *Tx) Begin(ctx context.Context) (*Tx, error) {
	savepoint := fmt.Sprintf("savepoint_%d", len(t.savepoints))
	t.savepoints = append(t.savepoints, savepoint)
	_, err := t.tx.Exec(ctx, `SAVEPOINT `+savepoint)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: t.tx, savepoints: t.savepoints, Queries: t.Queries}, nil
}

func (t *Tx) Commit(ctx context.Context) error {
	if len(t.savepoints) == 0 {
		return t.tx.Commit(ctx)
	}
	savepoint := t.savepoints[len(t.savepoints)-1]
	t.savepoints = t.savepoints[:len(t.savepoints)-1]
	_, err := t.tx.Exec(ctx, `RELEASE SAVEPOINT `+savepoint)
	return err
}

func (t *Tx) Rollback(ctx context.Context) error {
	if len(t.savepoints) == 0 {
		return t.tx.Rollback(ctx)
	}
	savepoint := t.savepoints[len(t.savepoints)-1]
	t.savepoints = t.savepoints[:len(t.savepoints)-1]
	_, err := t.tx.Exec(ctx, `ROLLBACK TO SAVEPOINT `+savepoint)
	return err
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
		*err = errors.Join(*err, t.Rollback(ctx))
	} else {
		*err = t.Commit(ctx)
	}
}
