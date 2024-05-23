package ftltest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/alecthomas/assert/v2"
)

var (
	keys1  = []string{"one", "1"}
	keys2  = []string{"two", "2"}
	module = "test"
)

func TestDoubleAcquireLease(t *testing.T) {
	ctx := context.Background()
	client := newFakeLeaseClient()

	// Acquire a lease, and immediately try to acquire it again.
	err := client.Acquire(ctx, module, keys1, 1*time.Second)
	assert.NoError(t, err)
	err = client.Acquire(ctx, module, keys1, 1*time.Second)
	assert.True(t, errors.Is(err, ftl.ErrLeaseHeld), "expected lease to already be held")
}

func TestAcquireTwoDifferentLeases(t *testing.T) {
	ctx := context.Background()
	client := newFakeLeaseClient()

	err := client.Acquire(ctx, module, keys1, 1*time.Second)
	assert.NoError(t, err)
	err = client.Acquire(ctx, module, keys2, 1*time.Second)
	assert.NoError(t, err)
}

func TestExpiry(t *testing.T) {
	ctx := context.Background()
	client := newFakeLeaseClient()

	err := client.Acquire(ctx, module, keys1, 500*time.Millisecond)
	assert.NoError(t, err)
	time.Sleep(250 * time.Millisecond)
	err = client.Heartbeat(ctx, module, keys1, 500*time.Millisecond)
	assert.NoError(t, err)
	time.Sleep(250 * time.Millisecond)
	err = client.Heartbeat(ctx, module, keys1, 500*time.Millisecond)
	assert.NoError(t, err)

	// wait longer than ttl
	time.Sleep(1 * time.Second)
	err = client.Heartbeat(ctx, module, keys1, 500*time.Millisecond)
	assert.Error(t, err, "expected error for heartbeating expired lease")
	err = client.Release(ctx, keys1)
	assert.Error(t, err, "expected error for heartbeating expired lease")

	// try and acquire again
	err = client.Acquire(ctx, module, keys1, 1*time.Second)
	assert.NoError(t, err)
}
