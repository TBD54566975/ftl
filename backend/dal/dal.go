package dal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
)

// Counters for testing.
var (
	testCommitCounter   atomic.Int64
	testRollbackCounter atomic.Int64
)

// Connection is a common interface for *sql.DB and *sql.Tx.
type Connection interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// MakeWithHandle is a function that can be used to create a new T with a SQLHandle.
type MakeWithHandle[T any] func(*Handle[T]) *T

// Handle is a wrapper around a database connection that can be embedded within a struct to provide access
// to a database connection and methods for managing transactions.
type Handle[T any] struct {
	Connection Connection
	txCounter  int64
	Make       MakeWithHandle[T]
}

// New creates a new Handle
func New[T any](sql Connection, fn MakeWithHandle[T]) *Handle[T] {
	return &Handle[T]{Connection: sql, Make: fn}
}

// Begin creates a new transaction or increments the transaction counter if the handle is already in a transaction.
//
// In all cases a new handle is returned.
func (h *Handle[T]) Begin(ctx context.Context) (*T, error) {
	switch conn := h.Connection.(type) {
	case *sql.DB:
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", TranslatePGError(err))
		}

		txConn := &Handle[T]{Connection: tx, Make: h.Make}
		return h.Make(txConn), nil

	case *sql.Tx:
		sub := &Handle[T]{Connection: conn, Make: h.Make, txCounter: h.txCounter + 1}
		_, err := conn.ExecContext(ctx, fmt.Sprintf("SAVEPOINT sp%d", sub.txCounter))
		if err != nil {
			return nil, fmt.Errorf("failed to begin savepoint: %w", TranslatePGError(err))
		}

		return h.Make(sub), nil
	default:
		return nil, errors.New("invalid connection type")
	}
}

// CommitOrRollback commits the transaction if err is nil, otherwise rolls it back.
//
// Use it in a defer like so, particularly taking note of named return value
// `(err error)`. Without this it will not work.
//
//	func (d *DAL) SomeMethod() (err error) {
//		tx, err := d.Begin()
//		if err != nil { return err }
//		defer tx.CommitOrRollback(&err)
//		// ...
//		return nil
//	}
func (h *Handle[T]) CommitOrRollback(ctx context.Context, err *error) {
	_, ok := h.Connection.(*sql.Tx)
	if !ok {
		*err = errors.New("can only commit or rollback a transaction")
		return
	}

	if h.txCounter > 0 {
		return
	}

	if *err != nil {
		*err = errors.Join(*err, h.Rollback(ctx))
	} else {
		*err = h.Commit(ctx)
	}
}

// Commit the transaction or savepoint.
func (h *Handle[T]) Commit(ctx context.Context) error {
	sqlTx, ok := h.Connection.(*sql.Tx)
	if !ok {
		return errors.New("can only commit or rollback a transaction")
	}

	testCommitCounter.Add(1)
	if h.txCounter == 0 {
		return TranslatePGError(sqlTx.Commit())
	}
	_, err := sqlTx.Exec(fmt.Sprintf("RELEASE SAVEPOINT sp%d", h.txCounter))
	if err != nil {
		return TranslatePGError(err)
	}
	return nil
}

// Rollback the transaction or savepoint.
func (h *Handle[T]) Rollback(ctx context.Context) error {
	sqlTx, ok := h.Connection.(*sql.Tx)
	if !ok {
		return errors.New("can only commit or rollback a transaction")
	}

	testRollbackCounter.Add(1)
	if h.txCounter == 0 {
		return TranslatePGError(sqlTx.Rollback())
	}
	_, err := sqlTx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT sp%d", h.txCounter))
	if err != nil {
		return TranslatePGError(err)
	}
	return nil
}
