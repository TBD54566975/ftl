package sqltest

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/internal/flock"
)

// OpenForTesting opens a database connection for testing, recreating the
// database beforehand.
func OpenForTesting(ctx context.Context, t testing.TB) *pgxpool.Pool {
	t.Helper()
	// Acquire lock for this DB.
	lockPath := filepath.Join(os.TempDir(), "ftl-db-test.lock")
	release, err := flock.Acquire(ctx, lockPath, 10*time.Second)
	assert.NoError(t, err)
	t.Cleanup(func() { _ = release() })

	testDSN := "postgres://localhost:15432/ftl-test?user=postgres&password=secret&sslmode=disable"
	conn, err := databasetesting.CreateForDevel(ctx, testDSN, true)
	assert.NoError(t, err)
	return conn
}
