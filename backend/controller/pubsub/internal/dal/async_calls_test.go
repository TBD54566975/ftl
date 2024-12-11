package dal

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

func TestNoCallToAcquire(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)

	timelineEndpoint, err := url.Parse("http://localhost:8080")
	assert.NoError(t, err)

	timelineClient := timeline.NewClient(ctx, timelineEndpoint)

	dal := New(conn, timelineClient)

	_, _, err = dal.AcquireAsyncCall(ctx)
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
