package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type DBI interface {
	Querier
	Conn() ConnI
	Begin(ctx context.Context) (*Tx, error)
}

type ConnI interface {
	DBTX
	Begin() (*sql.Tx, error)
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
	tx, err := d.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, Queries: New(tx)}, nil
}

type noopSubConn struct {
	DBTX
}

func (noopSubConn) Begin() (*sql.Tx, error) {
	return nil, errors.New("sql: not implemented")
}

type Tx struct {
	tx *sql.Tx
	*Queries
}

func (t *Tx) Conn() ConnI { return noopSubConn{t.tx} }

func (t *Tx) Tx() *sql.Tx { return t.tx }

func (t *Tx) Begin(ctx context.Context) (*Tx, error) {
	return nil, fmt.Errorf("cannot nest transactions")
}

func (t *Tx) Commit(ctx context.Context) error {
	err := t.tx.Commit()
	if err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (t *Tx) Rollback(ctx context.Context) error {
	err := t.tx.Rollback()
	if err != nil {
		return fmt.Errorf("rolling back transaction: %w", err)
	}

	return nil
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
