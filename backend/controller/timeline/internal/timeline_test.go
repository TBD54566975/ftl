package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/TBD54566975/ftl/backend/libdal"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/cronjobs"
	timeline2 "github.com/TBD54566975/ftl/backend/controller/timeline"

	controllerdal "github.com/TBD54566975/ftl/backend/controller/dal"
	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
)

func TestTimeline(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	timeline := timeline2.New(ctx, conn, encryption)
	registry := artefacts.New(artefacts.ContainerConfig{}, conn)
	pubSub := pubsub.New(ctx, conn, encryption, optional.None[pubsub.AsyncCallListener]())

	key := model.NewControllerKey("localhost", strconv.Itoa(8080+1))
	cjs := cronjobs.New(ctx, key, "test.com", encryption, timeline, conn)
	controllerDAL := controllerdal.New(ctx, conn, encryption, pubSub, cjs, func(c libdal.Connection) artefacts.Service {
		return nil
	})

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)

	t.Run("UpsertModule", func(t *testing.T) {
		err = controllerDAL.UpsertModule(ctx, "go", "test")
		assert.NoError(t, err)
	})

	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = registry.Upload(ctx, artefacts.Artefact{Content: testContent})
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = controllerDAL.CreateDeployment(ctx, "go", module, []dalmodel.DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil)
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

	callEvent := &timeline2.CallEvent{
		Time:          time.Now().Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time":"now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}

	t.Run("InsertCallEvent", func(t *testing.T) {
		call := timeline2.CallEventToCallForTesting(callEvent)
		timeline.EnqueueEvent(ctx, call)
		time.Sleep(200 * time.Millisecond)
	})

	logEvent := &timeline2.LogEvent{
		Log: timeline2.Log{
			Time:          time.Now().Round(time.Millisecond),
			DeploymentKey: deploymentKey,
			RequestKey:    optional.Some(requestKey),
			Level:         int32(log.Warn),
			Attributes:    map[string]string{"attr": "value"},
			Message:       "A log entry",
		},
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		timeline.EnqueueEvent(ctx, &logEvent.Log)
		time.Sleep(200 * time.Millisecond)
	})

	ingressEvent := &timeline2.IngressEvent{
		DeploymentKey:  deploymentKey,
		RequestKey:     optional.Some(requestKey),
		Verb:           schema.Ref{Module: "echo", Name: "echo"},
		Method:         "GET",
		Path:           "/echo",
		StatusCode:     200,
		Time:           time.Now().Round(time.Millisecond),
		Request:        []byte(`{"request":"body"}`),
		RequestHeader:  json.RawMessage(`{"request":["header"]}`),
		Response:       []byte(`{"response":"body"}`),
		ResponseHeader: json.RawMessage(`{"response":["header"]}`),
	}

	t.Run("InsertHTTPIngressEvent", func(t *testing.T) {
		timeline.EnqueueEvent(ctx, &timeline2.Ingress{
			DeploymentKey:   ingressEvent.DeploymentKey,
			RequestKey:      ingressEvent.RequestKey.MustGet(),
			StartTime:       ingressEvent.Time,
			Verb:            &ingressEvent.Verb,
			RequestMethod:   ingressEvent.Method,
			RequestPath:     ingressEvent.Path,
			RequestHeaders:  http.Header(map[string][]string{"request": {"header"}}),
			RequestBody:     ingressEvent.Request,
			ResponseStatus:  ingressEvent.StatusCode,
			ResponseHeaders: http.Header(map[string][]string{"response": {"header"}}),
			ResponseBody:    ingressEvent.Response,
		})
		time.Sleep(200 * time.Millisecond)
	})

	cronEvent := &timeline2.CronScheduledEvent{
		CronScheduled: timeline2.CronScheduled{
			DeploymentKey: deploymentKey,
			Verb:          schema.Ref{Module: "time", Name: "time"},
			Time:          time.Now().Round(time.Millisecond),
			ScheduledAt:   time.Now().Add(time.Minute).Round(time.Millisecond).UTC(),
			Schedule:      "* * * * *",
			Error:         optional.None[string](),
		},
	}

	t.Run("InsertCronScheduledEvent", func(t *testing.T) {
		timeline.EnqueueEvent(ctx, &timeline2.CronScheduled{
			DeploymentKey: cronEvent.DeploymentKey,
			Verb:          cronEvent.Verb,
			Time:          cronEvent.Time,
			ScheduledAt:   cronEvent.ScheduledAt,
			Schedule:      cronEvent.Schedule,
			Error:         cronEvent.Error,
		})
		time.Sleep(200 * time.Millisecond)
	})

	expectedDeploymentUpdatedEvent := &timeline2.DeploymentUpdatedEvent{
		DeploymentKey: deploymentKey,
		MinReplicas:   1,
	}

	t.Run("QueryEvents", func(t *testing.T) {
		t.Run("Limit", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(events))
		})

		t.Run("NoFilters", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000)
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{expectedDeploymentUpdatedEvent, callEvent, logEvent, ingressEvent, cronEvent}, events)
		})

		t.Run("ByDeployment", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterDeployments(deploymentKey))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{expectedDeploymentUpdatedEvent, callEvent, logEvent, ingressEvent, cronEvent}, events)
		})

		t.Run("ByCall", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterTypes(timeline2.EventTypeCall), timeline2.FilterCall(optional.None[string](), "time", optional.None[string]()))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{callEvent}, events)
		})

		t.Run("ByModule", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterTypes(timeline2.EventTypeIngress), timeline2.FilterModule("echo", optional.None[string]()))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{ingressEvent}, events)
		})

		t.Run("ByModuleWithVerb", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterTypes(timeline2.EventTypeIngress), timeline2.FilterModule("echo", optional.Some("echo")))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{ingressEvent}, events)
		})

		t.Run("ByLogLevel", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterTypes(timeline2.EventTypeLog), timeline2.FilterLogLevel(log.Trace))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{logEvent}, events)
		})

		t.Run("ByRequests", func(t *testing.T) {
			events, err := timeline.QueryTimeline(ctx, 1000, timeline2.FilterRequests(requestKey))
			assert.NoError(t, err)
			assertEventsEqual(t, []timeline2.Event{callEvent, logEvent, ingressEvent}, events)
		})
	})
}

func normaliseEvents(events []timeline2.Event) []timeline2.Event {
	for i := range events {
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

func assertEventsEqual(t *testing.T, expected, actual []timeline2.Event) {
	t.Helper()
	assert.Equal(t, normaliseEvents(expected), normaliseEvents(actual), assert.Exclude[time.Duration](), assert.Exclude[time.Time]())
}

func TestDeleteOldEvents(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	encryption, err := encryption.New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	timeline := timeline2.New(ctx, conn, encryption)
	registry := artefacts.New(artefacts.ContainerConfig{}, conn)
	pubSub := pubsub.New(ctx, conn, encryption, optional.None[pubsub.AsyncCallListener]())
	controllerDAL := controllerdal.New(ctx, conn, encryption, pubSub, nil, func(c libdal.Connection) artefacts.Service {
		return nil
	})

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = registry.Upload(ctx, artefacts.Artefact{Content: testContent})
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = controllerDAL.CreateDeployment(ctx, "go", module, []dalmodel.DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil)
		assert.NoError(t, err)
	})

	requestKey := model.NewRequestKey(model.OriginIngress, "GET /test")
	// week old event
	callEvent := &timeline2.CallEvent{
		Time:          time.Now().Add(-24 * 7 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time": "now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}

	t.Run("InsertCallEvent", func(t *testing.T) {
		call := timeline2.CallEventToCallForTesting(callEvent)
		timeline.EnqueueEvent(ctx, call)
		time.Sleep(200 * time.Millisecond)
	})
	// hour old event
	callEvent = &timeline2.CallEvent{
		Time:          time.Now().Add(-1 * time.Hour).Round(time.Millisecond),
		DeploymentKey: deploymentKey,
		RequestKey:    optional.Some(requestKey),
		Request:       []byte("{}"),
		Response:      []byte(`{"time": "now"}`),
		DestVerb:      schema.Ref{Module: "time", Name: "time"},
	}
	t.Run("InsertCallEvent", func(t *testing.T) {
		call := timeline2.CallEventToCallForTesting(callEvent)
		timeline.EnqueueEvent(ctx, call)
		time.Sleep(200 * time.Millisecond)
	})

	// week old event
	logEvent := &timeline2.LogEvent{
		Log: timeline2.Log{
			Time:          time.Now().Add(-24 * 7 * time.Hour).Round(time.Millisecond),
			DeploymentKey: deploymentKey,
			RequestKey:    optional.Some(requestKey),
			Level:         int32(log.Warn),
			Attributes:    map[string]string{"attr": "value"},
			Message:       "A log entry",
		},
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		timeline.EnqueueEvent(ctx, &logEvent.Log)
		time.Sleep(200 * time.Millisecond)
	})

	// hour old event
	logEvent = &timeline2.LogEvent{
		Log: timeline2.Log{
			Time:          time.Now().Add(-1 * time.Hour).Round(time.Millisecond),
			DeploymentKey: deploymentKey,
			RequestKey:    optional.Some(requestKey),
			Level:         int32(log.Warn),
			Attributes:    map[string]string{"attr": "value"},
			Message:       "A log entry",
		},
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		timeline.EnqueueEvent(ctx, &logEvent.Log)
		time.Sleep(200 * time.Millisecond)
	})

	t.Run("DeleteOldEvents", func(t *testing.T) {
		count, err := timeline.DeleteOldEvents(ctx, timeline2.EventTypeCall, 2*24*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		count, err = timeline.DeleteOldEvents(ctx, timeline2.EventTypeLog, time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		count, err = timeline.DeleteOldEvents(ctx, timeline2.EventTypeLog, time.Minute)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
