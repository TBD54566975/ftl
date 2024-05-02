package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

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

	ref := schema.Ref{Module: "module", Name: "verb"}
	err = dal.SendFSMEvent(ctx, "test", "invoiceID", "state", ref, []byte(`{}`))
	assert.NoError(t, err)

	call, err := dal.AcquireAsyncCall(ctx)
	assert.NoError(t, err)
	t.Cleanup(func() {
		err := call.Lease.Release()
		assert.NoError(t, err)
	})

	assert.HasPrefix(t, call.Lease.String(), "/system/async_call/1:")
	expectedCall := &AsyncCall{
		ID:        1,
		Origin:    AsyncCallOriginFSM,
		OriginKey: "invoiceID",
		Verb:      ref,
		Request:   []byte(`{}`),
	}
	assert.Equal(t, expectedCall, call, assert.Exclude[*Lease]())

	err = dal.CompleteAsyncCall(ctx, call, nil, optional.None[string]())
	assert.EqualError(t, err, "must provide exactly one of response or error")

	err = dal.CompleteAsyncCall(ctx, call, []byte(`{}`), optional.None[string]())
	assert.NoError(t, err)

	actual, err := dal.LoadAsyncCall(ctx, call.ID)
	assert.NoError(t, err)
	assert.Equal(t, call, actual, assert.Exclude[*Lease]())
}
