package sqltest

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/gofrs/flock"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
)

// OpenForTesting opens a database connection for testing, recreating the
// database beforehand.
func OpenForTesting(ctx context.Context, t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Acquire lock for this DB.
	lockPath := filepath.Join(os.TempDir(), "ftl-db-test.lock")
	lock := flock.New(lockPath)
	ok, err := lock.TryLockContext(ctx, time.Second)
	assert.NoError(t, err)
	assert.True(t, ok, "could not acquire lock on %s", lockPath)
	t.Cleanup(func() { _ = lock.Unlock() })

	testDSN := "postgres://localhost:54320/ftl-test?user=postgres&password=secret&sslmode=disable"
	conn, err := databasetesting.CreateForDevel(ctx, testDSN, true)
	assert.NoError(t, err)
	return conn
}
