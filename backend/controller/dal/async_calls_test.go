package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/routing"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

func TestNoCallToAcquire(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)

	pubSub := pubsub.New(ctx, conn, optional.None[pubsub.AsyncCallListener](), routing.New(ctx, schemaeventsource.NewUnattached()))
	dal := New(ctx, conn, pubSub, nil)

	_, _, err := dal.AcquireAsyncCall(ctx)
	assert.IsError(t, err, libdal.ErrNotFound)
	assert.EqualError(t, err, "no pending async calls: not found")
}

func TestParser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected async.AsyncOrigin
	}{
		{"Cron", `cron:crn-initial-verb0Cron-5eq2ivpmuv0lvnoc`, async.AsyncOriginCron{
			CronJobKey: model.CronJobKey{
				Payload: model.CronJobPayload{Module: "initial", Verb: "verb0Cron"},
				Suffix:  []byte("\xfd7\xe6*\xfcƹ\xe9.\x9c"),
			}}},
		{"PubSub", `sub:module.topic`, async.AsyncOriginPubSub{Subscription: schema.RefKey{Module: "module", Name: "topic"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			origin, err := async.ParseAsyncOrigin(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, origin)
		})
	}
}
