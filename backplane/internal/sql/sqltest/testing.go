package sqltest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/gofrs/flock"
	"github.com/jackc/pgx/v5"

	"github.com/TBD54566975/ftl/backplane/internal/sql"
)

const testDatabaseName = "ftl-test"

// OpenForTesting opens a database connection for testing, recreating the
// database beforehand.
func OpenForTesting(t *testing.T) sql.DBI {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Acquire lock for this DB.
	lockPath := filepath.Join(os.TempDir(), "ftl-db-test.lock")
	lock := flock.New(lockPath)
	ok, err := lock.TryLockContext(ctx, time.Second)
	assert.NoError(t, err)
	assert.True(t, ok, "could not acquire lock on %s", lockPath)
	t.Cleanup(func() { _ = lock.Unlock() })

	conn, err := pgx.Connect(ctx, makeTestDSN("postgres"))
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Close(ctx)
	})

	// Terminate any dangling connections.
	_, err = conn.Exec(ctx, `
		SELECT pid, pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()`,
		testDatabaseName)
	assert.NoError(t, err)

	_, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", testDatabaseName))
	assert.NoError(t, err)
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %q", testDatabaseName))
	assert.NoError(t, err)
	dsn := makeTestDSN(testDatabaseName)

	err = sql.Migrate(ctx, dsn)
	assert.NoError(t, err)

	realConn, err := pgx.Connect(ctx, dsn)
	assert.NoError(t, err)
	return realConn
}

func makeTestDSN(database string) string {
	return fmt.Sprintf("postgres://localhost/%s?user=postgres&password=secret&sslmode=disable", database)
}
