package leases

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestFakeLease(t *testing.T) {
	leaser := NewFakeLeaser()

	lease1, err := leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second)
	assert.NoError(t, err)

	_, err = leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second)
	assert.IsError(t, err, ErrConflict)

	err = lease1.Release()
	assert.NoError(t, err)

	lease2, err := leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second)
	assert.NoError(t, err)
	err = lease2.Release()
	assert.NoError(t, err)
}
