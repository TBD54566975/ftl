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
