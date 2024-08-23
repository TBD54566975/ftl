package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	dalerrs "github.com/TBD54566975/ftl/backend/dal"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestNoCallToAcquire(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	_, err = dal.AcquireAsyncCall(ctx)
	assert.IsError(t, err, dalerrs.ErrNotFound)
	assert.EqualError(t, err, "no pending async calls: not found")
}

func TestParseAsyncOrigin(t *testing.T) {
	cronKeys := []string{
		"crn-cron-cron-10pvs393nkv3new4",         // 1:23: exponent has no digits
		"crn-initial-verb0-3poj0hr6wmtvmz99",     // 1:26: exponent has no digits
		"crn-initial-verb0Cron-5eq2ivpmuv0lvnoc", // 1:30: exponent has no digits
	}
	for _, cronKey := range cronKeys {
		origin, err := ParseAsyncOrigin("cron:" + cronKey)
		assert.NoError(t, err)
		assert.Equal(t, "cron", origin.Origin())

		asyncOrigin, ok := origin.(*AsyncOriginCron)
		assert.True(t, ok, "origin is not a cron origin")
		assert.Equal(t, cronKey, asyncOrigin.CronJobKey)
	}
}
