package dal

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	controllerdal "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	ftlencryption "github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

func TestTimelineDAL(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	encryption, err := encryption.New(ctx, conn, ftlencryption.NewBuilder())
	assert.NoError(t, err)

	dal := New(conn, encryption)
	controllerDAL := controllerdal.New(ctx, conn, encryption)

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)

	t.Run("UpsertModule", func(t *testing.T) {
		err = controllerDAL.UpsertModule(ctx, "go", "test")
		assert.NoError(t, err)
	})

	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = controllerDAL.CreateArtefact(ctx, testContent)
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = controllerDAL.CreateDeployment(ctx, "go", module, []controllerdal.DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := controllerDAL.SetDeploymentReplicas(ctx, deploymentKey, 1)
		assert.NoError(t, err)
	})

	requestKey := model.NewRequestKey(model.OriginIngress, "GET /test")
	t.Run("CreateIngressRequest", func(t *testing.T) {
		err = controllerDAL.CreateRequest(ctx, requestKey, "127.0.0.1:1234")
		assert.NoError(t, err)
	})

	callEvent := &CallEvent{
		Time:          time.Now().Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time":"now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}
	t.Run("InsertCallEvent", func(t *testing.T) {
		err = dal.InsertCallEvent(ctx, callEvent)
		assert.NoError(t, err)
	})

	logEvent := &LogEvent{
		Time:          time.Now().Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Level:         int32(log.Warn),
		Attributes:    map[string]string{"attr": "value"},
		Message:       "A log entry",
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		err = dal.InsertLogEvent(ctx, logEvent)
		assert.NoError(t, err)
	})

	expectedDeploymentUpdatedEvent := &DeploymentUpdatedEvent{
		DeploymentKey: deploymentKey,
		MinReplicas:   1,
	}

	t.Run("QueryEvents", func(t *testing.T) {
		t.Run("Limit", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(events))
		})

		t.Run("NoFilters", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1000)
			assert.NoError(t, err)
			assertEventsEqual(t, []TimelineEvent{expectedDeploymentUpdatedEvent, callEvent, logEvent}, events)
		})

		t.Run("ByDeployment", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1000, FilterDeployments(deploymentKey))
			assert.NoError(t, err)
			assertEventsEqual(t, []TimelineEvent{expectedDeploymentUpdatedEvent, callEvent, logEvent}, events)
		})

		t.Run("ByCall", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1000, FilterTypes(EventTypeCall), FilterCall(optional.None[string](), "time", optional.None[string]()))
			assert.NoError(t, err)
			assertEventsEqual(t, []TimelineEvent{callEvent}, events)
		})

		t.Run("ByLogLevel", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1000, FilterTypes(EventTypeLog), FilterLogLevel(log.Trace))
			assert.NoError(t, err)
			assertEventsEqual(t, []TimelineEvent{logEvent}, events)
		})

		t.Run("ByRequests", func(t *testing.T) {
			events, err := dal.QueryTimeline(ctx, 1000, FilterRequests(requestKey))
			assert.NoError(t, err)
			assertEventsEqual(t, []TimelineEvent{callEvent, logEvent}, events)
		})
	})
}

func normaliseEvents(events []TimelineEvent) []TimelineEvent {
	for i := range len(events) {
		event := events[i]
		re := reflect.Indirect(reflect.ValueOf(event))
		f := re.FieldByName("Time")
		f.Set(reflect.Zero(f.Type()))
		f = re.FieldByName("ID")
		f.Set(reflect.Zero(f.Type()))
		events[i] = event
	}

	return events
}

func assertEventsEqual(t *testing.T, expected, actual []TimelineEvent) {
	t.Helper()
	assert.Equal(t, normaliseEvents(expected), normaliseEvents(actual))
}

func TestDeleteOldEvents(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	encryption, err := encryption.New(ctx, conn, ftlencryption.NewBuilder())
	assert.NoError(t, err)

	dal := New(conn, encryption)
	controllerDAL := controllerdal.New(ctx, conn, encryption)

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = controllerDAL.CreateArtefact(ctx, testContent)
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = controllerDAL.CreateDeployment(ctx, "go", module, []controllerdal.DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil, nil)
		assert.NoError(t, err)
	})

	requestKey := model.NewRequestKey(model.OriginIngress, "GET /test")
	// week old event
	callEvent := &CallEvent{
		Time:          time.Now().Add(-24 * 7 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time": "now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}
	t.Run("InsertCallEvent", func(t *testing.T) {
		err = dal.InsertCallEvent(ctx, callEvent)
		assert.NoError(t, err)
	})
	// hour old event
	callEvent = &CallEvent{
		Time:          time.Now().Add(-1 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time": "now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}
	t.Run("InsertCallEvent", func(t *testing.T) {
		err = dal.InsertCallEvent(ctx, callEvent)
		assert.NoError(t, err)
	})

	// week old event
	logEvent := &LogEvent{
		Time:          time.Now().Add(-24 * 7 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Level:         int32(log.Warn),
		Attributes:    map[string]string{"attr": "value"},
		Message:       "A log entry",
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		err = dal.InsertLogEvent(ctx, logEvent)
		assert.NoError(t, err)
	})
	// hour old event
	logEvent = &LogEvent{
		Time:          time.Now().Add(-1 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Level:         int32(log.Warn),
		Attributes:    map[string]string{"attr": "value"},
		Message:       "A log entry",
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		err = dal.InsertLogEvent(ctx, logEvent)
		assert.NoError(t, err)
	})

	t.Run("DeleteOldEvents", func(t *testing.T) {
		count, err := dal.DeleteOldEvents(ctx, EventTypeCall, 2*24*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		count, err = dal.DeleteOldEvents(ctx, EventTypeLog, time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = dal.DeleteOldEvents(ctx, EventTypeLog, time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
