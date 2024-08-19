package sqltest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/flock"
)

// OpenForTesting opens a database connection for testing, recreating the
// database beforehand.
func OpenForTesting(ctx context.Context, t testing.TB) *sql.DB {
	t.Helper()
	// Acquire lock for this DB.
	lockPath := filepath.Join(os.TempDir(), "ftl-db-test.lock")
	release, err := flock.Acquire(ctx, lockPath, 30*time.Second)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = release() }) //nolint:errcheck

	testDSN := "postgres://localhost:15432/ftl-test?user=postgres&password=secret&sslmode=disable"
	conn, err := databasetesting.CreateForDevel(ctx, testDSN, true)
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })
	return conn
}
