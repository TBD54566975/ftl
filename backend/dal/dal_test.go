package dal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
	_ "modernc.org/sqlite" // Pure Go SQLite driver.
)

type DAL struct {
	*Handle[DAL]
}

// New creates a new Data Access Layer instance.
func NewConn(sqlConn *sql.DB) *DAL {
	return NewWithConn(New(sqlConn, NewWithConn))
}

func NewWithConn(conn *Handle[DAL]) *DAL {
	return &DAL{conn}
}

func (d *DAL) CreateUser(ctx context.Context, username string, name string) error {
	_, err := d.Connection.ExecContext(ctx, `
		INSERT INTO users (username, name)
		VALUES ($1, $2)
		`, username, name)
	if err != nil {
		return fmt.Errorf("create user %s: %w", username, err)
	}
	return nil
}

func (d *DAL) CreateUsers(ctx context.Context, users [][]string) (err error) {
	txn, err := d.Begin(ctx)
	if err != nil {
		return err
	}

	defer txn.CommitOrRollback(ctx, &err)

	for _, user := range users {
		err = txn.CreateUser(ctx, user[0], user[1])
		if err != nil {
			return err
		}
	}

	return err
}

func (d *DAL) GetUserByUsername(ctx context.Context, username string) (string, error) {
	var user string
	err := d.Connection.QueryRowContext(ctx, `
		SELECT name
		FROM users
		WHERE username = $1
		`, username).Scan(&user)
	if err != nil {
		return user, fmt.Errorf("user by username %s: %w", username, err)
	}
	return user, nil
}

func TestDAL(t *testing.T) {
	for _, test := range []struct {
		name string
		fn   func(ctx context.Context, t *testing.T, dal *DAL)
	}{
		{"WriteAndRead", func(ctx context.Context, t *testing.T, dal *DAL) {
			err := dal.CreateUser(ctx, "bob", "Bob Smith")
			assert.NoError(t, err)

			user, err := dal.GetUserByUsername(ctx, "bob")
			assert.NoError(t, err)
			assert.Equal(t, "Bob Smith", user)
		}},
		{"CommitOrRollbackWillRollbackOnError", func(ctx context.Context, t *testing.T, dal *DAL) {
			f := func() (err error) {
				tx, err := dal.Begin(ctx)
				assert.NoError(t, err)
				defer tx.CommitOrRollback(ctx, &err)

				err = tx.CreateUser(ctx, "bob", "Bob Smith")
				assert.NoError(t, err)

				return errors.New("some error")
			}

			err := f()
			assert.EqualError(t, err, "some error")
			assert.Equal(t, 1, testRollbackCounter.Load())
			assert.Equal(t, 0, testCommitCounter.Load())

			_, err = dal.GetUserByUsername(ctx, "bob")
			assert.IsError(t, err, sql.ErrNoRows)
		}},
		{"CommitOrRollbackWillCommitOnSuccess", func(ctx context.Context, t *testing.T, dal *DAL) {
			f := func() (err error) {
				tx, err := dal.Begin(ctx)
				assert.NoError(t, err)
				defer tx.CommitOrRollback(ctx, &err)

				err = tx.CreateUser(ctx, "bob", "Bob Smith")
				assert.NoError(t, err)

				return nil
			}

			err := f()
			assert.NoError(t, err)
			assert.Equal(t, 0, testRollbackCounter.Load())
			assert.Equal(t, 1, testCommitCounter.Load())

			user, err := dal.GetUserByUsername(ctx, "bob")
			assert.NoError(t, err)
			assert.Equal(t, "Bob Smith", user)
		}},
		{"TestMultipleTxn", func(ctx context.Context, t *testing.T, dal *DAL) {
			f := func() (err error) {
				tx, err := dal.Begin(ctx)
				assert.NoError(t, err)
				defer tx.CommitOrRollback(ctx, &err)

				err = tx.CreateUser(ctx, "bob", "Bob Smith")
				assert.NoError(t, err)

				err = tx.CreateUsers(ctx, [][]string{
					{"randy", "Randy McRando"},
					{"hehe", "Jimmy DROP TABLES"},
				})

				assert.NoError(t, err)

				return nil
			}

			err := f()
			assert.NoError(t, err)
			assert.Equal(t, 0, testRollbackCounter.Load())
			assert.Equal(t, 1, testCommitCounter.Load())

			user, err := dal.GetUserByUsername(ctx, "bob")
			assert.NoError(t, err)
			assert.Equal(t, "Bob Smith", user)

			user2, err := dal.GetUserByUsername(ctx, "randy")
			assert.NoError(t, err)
			assert.Equal(t, "Randy McRando", user2)

			user3, err := dal.GetUserByUsername(ctx, "hehe")
			assert.NoError(t, err)
			assert.Equal(t, "Jimmy DROP TABLES", user3)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				testRollbackCounter.Store(0)
				testCommitCounter.Store(0)
			})
			ctx := context.Background()
			db, err := sql.Open("sqlite", ":memory:")
			assert.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, db.Close()) })
			_, err = db.Exec(`CREATE TABLE users (username TEXT, name TEXT)`)
			assert.NoError(t, err)
			test.fn(ctx, t, NewConn(db))
		})
	}
}
