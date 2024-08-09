package main

import (
	"context"
	"database/sql"

	libdal "pfi/backend/libs/dal"
)

type Tx struct {
	*libdal.Tx[Tx]
	*DAL
}

// DAL is the data access layer for the IDV module.
type DAL struct {
	db libdal.Conn
	*libdal.DAL[Tx]
}

// NewDAL creates a new Data Access Layer instance.
func NewDAL(conn *sql.DB) *DAL {
	return &DAL{db: conn, DAL: libdal.New(conn, func(tx *sql.Tx, t *libdal.Tx[Tx]) *Tx {
		return &Tx{DAL: &DAL{db: tx, DAL: t.DAL}, Tx: t}
	})}
}

func failure() error {
	_ = func() error {
		dal := DAL{}
		tx, err := dal.Begin(context.Background())
		if err != nil {
			return err
		}
		defer tx.CommitOrRollback(&err) // Should error
		return nil
	}

	dal := DAL{}
	tx, err := dal.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.CommitOrRollback(&err) // Should error
	return nil
}

func success() (err error) {
	_ = func() error {
		dal := DAL{}
		tx, err := dal.Begin(context.Background())
		if err != nil {
			return err
		}
		defer tx.CommitOrRollback(&err) // Should error
		return nil
	}
	dal := DAL{}
	tx, err := dal.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.CommitOrRollback(&err) // Should NOT error
	return nil
}

func main() {
	_ = failure()
	_ = success()
}
