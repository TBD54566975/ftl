package leases

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestFakeLease(t *testing.T) {
	leaser := NewFakeLeaser()

	lease1, lease1Ctx, err := leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second, optional.None[any]())
	assert.NoError(t, err)

	_, _, err = leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second, optional.None[any]())
	assert.IsError(t, err, ErrConflict)

	err = lease1.Release()
	assert.NoError(t, err)

	lease2, lease2Ctx, err := leaser.AcquireLease(context.Background(), SystemKey("test"), time.Second, optional.None[any]())
	assert.NoError(t, err)
	err = lease2.Release()
	assert.NoError(t, err)

	time.Sleep(time.Second)
	assert.Error(t, lease1Ctx.Err(), "context should be cancelled after lease1 was released")
	assert.Error(t, lease2Ctx.Err(), "context should be cancelled after lease2 was released")
}
