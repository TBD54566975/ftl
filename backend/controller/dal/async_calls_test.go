package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
)

func TestSendFSMEvent(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)

	_, err = dal.AcquireAsyncCall(ctx)
	assert.IsError(t, err, ErrNotFound)

	err = dal.SendFSMEvent(ctx, "test", "test", "state", schema.Ref{Module: "module", Name: "verb"}, []byte(`{}`))
	assert.NoError(t, err)

	lease, err := dal.AcquireAsyncCall(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() {
		err := lease.Release()
		assert.NoError(t, err)
	})

	assert.HasPrefix(t, lease.String(), "/system/async_call/1:")
}
