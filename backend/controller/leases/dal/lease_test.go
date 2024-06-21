package dal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/log"
)

func leaseExists(t *testing.T, conn sql.ConnI, idempotencyKey uuid.UUID, key leases.Key) bool {
	t.Helper()
	var count int
	err := dalerrs.TranslatePGError(conn.
		QueryRow(context.Background(), "SELECT COUNT(*) FROM leases WHERE idempotency_key = $1 AND key = $2", idempotencyKey, key).
		Scan(&count))
	if errors.Is(err, dalerrs.ErrNotFound) {
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
	dal := New(conn)

	// TTL is too short, expect an error
	_, _, err := dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*1, optional.None[any]())
	assert.Error(t, err)

	leasei, leaseCtx, err := dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5, optional.None[any]())
	assert.NoError(t, err)

	lease := leasei.(*Lease) //nolint:forcetypeassert

	// Try to acquire the same lease again, which should fail.
	_, _, err = dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5, optional.None[any]())
	assert.IsError(t, err, leases.ErrConflict)

	time.Sleep(time.Second * 6)

	assert.True(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	err = lease.Release()
	assert.NoError(t, err)

	assert.False(t, leaseExists(t, conn, lease.idempotencyKey, lease.key))

	time.Sleep(time.Second)
	assert.Error(t, leaseCtx.Err(), "context should be cancelled after lease was released")
}

func TestExpireLeases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal := New(conn)

	leasei, _, err := dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5, optional.None[any]())
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

	leasei, _, err = dal.AcquireLease(ctx, leases.SystemKey("test"), time.Second*5, optional.None[any]())
	assert.NoError(t, err)

	err = leasei.Release()
	assert.NoError(t, err)
}
