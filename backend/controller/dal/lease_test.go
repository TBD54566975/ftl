package dal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
)

func leaseExists(t *testing.T, conn sql.ConnI, idempotencyKey uuid.UUID, key leases.Key) bool {
	t.Helper()
	var count int
	err := translatePGError(conn.
		QueryRow(context.Background(), "SELECT COUNT(*) FROM leases WHERE idempotency_key = $1 AND key = $2", idempotencyKey, key).
		Scan(&count))
	if errors.Is(err, ErrNotFound) {
		return false
	}
	assert.NoError(t, err)
	return count > 0
}

func TestLease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)

	// TTL is too short, expect an error
	_, err = dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*1)
	assert.Error(t, err)

	leasei, err := dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5)
	assert.NoError(t, err)

	lease := leasei.(*Lease) //nolint:forcetypeassert

	// Try to acquire the same lease again, which should fail.
	_, err = dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5)
	assert.IsError(t, err, ErrConflict)

	time.Sleep(time.Second * 6)

	assert.True(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	err = lease.Release()
	assert.NoError(t, err)

	assert.False(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))
}

func TestExpireLeases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)

	leasei, err := dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5)
	assert.NoError(t, err)

	lease := leasei.(*Lease) //nolint:forcetypeassert

	err = dal.ExpireLeases(ctx)
	assert.NoError(t, err)

	assert.True(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	// Pretend that the lease expired.
	lease.leak = true
	err = lease.Release()
	assert.NoError(t, err)

	assert.True(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	time.Sleep(time.Second * 6)

	err = dal.ExpireLeases(ctx)
	assert.NoError(t, err)

	assert.False(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	leasei, err = dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5)
	assert.NoError(t, err)

	err = leasei.Release()
	assert.NoError(t, err)
}
