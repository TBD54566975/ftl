package dal

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

func TestNoCallToAcquire(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	dal := New(ctx, conn, encryption)

	_, _, err = dal.AcquireAsyncCall(ctx)
	assert.IsError(t, err, libdal.ErrNotFound)
	assert.EqualError(t, err, "no pending async calls: not found")
}

func TestParser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected AsyncOrigin
	}{
		{"Cron", `cron:crn-initial-verb0Cron-5eq2ivpmuv0lvnoc`, AsyncOriginCron{
			CronJobKey: model.CronJobKey{
				Payload: model.CronJobPayload{Module: "initial", Verb: "verb0Cron"},
				Suffix:  []byte("\xfd7\xe6*\xfcƹ\xe9.\x9c"),
			}}},
		{"FSM", `fsm:module.name:key`, AsyncOriginFSM{FSM: schema.RefKey{Module: "module", Name: "name"}, Key: "key"}},
		{"PubSub", `sub:module.topic`, AsyncOriginPubSub{Subscription: schema.RefKey{Module: "module", Name: "topic"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			origin, err := asyncOriginParser.ParseString("", tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, origin.Key)
		})
	}
}
