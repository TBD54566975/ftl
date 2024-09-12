package dal

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/libdal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

//nolint:maintidx
func TestDAL(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSHA = sha256.Sum(testContent)

	deploymentChangesCh := dal.DeploymentChanges.Subscribe(nil)
	deploymentChanges := []DeploymentNotification{}
	wg := errgroup.Group{}
	wg.Go(func() error {
		for change := range deploymentChangesCh {
			deploymentChanges = append(deploymentChanges, change)
		}
		return nil
	})

	t.Run("UpsertModule", func(t *testing.T) {
		err = dal.UpsertModule(ctx, "go", "test")
		assert.NoError(t, err)
	})

	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = dal.CreateArtefact(ctx, testContent)
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = dal.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil, nil)
		assert.NoError(t, err)
	})

	deployment := &model.Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Key:      deploymentKey,
		Artefacts: []*model.Artefact{
			{Path: "dir/filename",
				Executable: true,
				Digest:     testSHA,
				Content:    io.NopCloser(bytes.NewReader(testContent))},
		},
	}
	expectedContent := artefactContent(t, deployment.Artefacts)

	t.Run("GetDeployment", func(t *testing.T) {
		actual, err := dal.GetDeployment(ctx, deploymentKey)
		assert.NoError(t, err)
		actualContent := artefactContent(t, actual.Artefacts)
		assert.Equal(t, expectedContent, actualContent)
		assert.Equal(t, deployment, actual)
	})

	t.Run("GetMissingDeployment", func(t *testing.T) {
		_, err := dal.GetDeployment(ctx, model.NewDeploymentKey("invalid"))
		assert.IsError(t, err, libdal.ErrNotFound)
	})

	t.Run("GetMissingArtefacts", func(t *testing.T) {
		misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
		missing, err := dal.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
		assert.NoError(t, err)
		assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
	})

	runnerID := model.NewRunnerKey("localhost", "8080")
	labels := map[string]any{"languages": []any{"go"}}

	t.Run("RegisterRunner", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateNew,
			Deployment: deploymentKey,
		})
		assert.NoError(t, err)
	})

	t.Run("RegisterRunnerFailsOnDuplicate", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:        model.NewRunnerKey("localhost", "8080"),
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateNew,
			Deployment: deploymentKey,
		})
		assert.Error(t, err)
		assert.IsError(t, err, libdal.ErrConflict)
	})

	expectedRunner := Runner{
		Key:        runnerID,
		Labels:     labels,
		Endpoint:   "http://localhost:8080",
		Deployment: deploymentKey,
	}

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := dal.SetDeploymentReplicas(ctx, deploymentKey, 1)
		assert.NoError(t, err)
	})

	t.Run("UpdateRunnerAssigned", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: deploymentKey,
		})
		assert.NoError(t, err)
	})

	t.Run("GetRunnersForDeployment", func(t *testing.T) {
		runners, err := dal.GetRunnersForDeployment(ctx, deploymentKey)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: deploymentKey,
		}}, runners)
	})

	requestKey := model.NewRequestKey(model.OriginIngress, "GET /test")
	t.Run("CreateIngressRequest", func(t *testing.T) {
		err = dal.CreateRequest(ctx, requestKey, "127.0.0.1:1234")
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

	t.Run("GetRoutingTable", func(t *testing.T) {
		routes, err := dal.GetRoutingTable(ctx, []string{deployment.Module})
		assert.NoError(t, err)
		assert.Equal(t, []Route{{
			Module:     "test",
			Runner:     expectedRunner.Key,
			Deployment: deploymentKey,
			Endpoint:   expectedRunner.Endpoint,
		}}, routes[deployment.Module])
	})

	t.Run("UpdateRunnerInvalidDeployment", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: model.NewDeploymentKey("test"),
		})
		assert.Error(t, err)
		assert.IsError(t, err, libdal.ErrConstraint)
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		_, err := dal.GetRoutingTable(ctx, []string{"non-existent"})
		assert.IsError(t, err, libdal.ErrNotFound)
	})

	t.Run("DeregisterRunner", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, runnerID)
		assert.NoError(t, err)
	})

	t.Run("DeregisterRunnerFailsOnMissing", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, model.NewRunnerKey("localhost", "8080"))
		assert.IsError(t, err, libdal.ErrNotFound)
	})

	t.Run("VerifyDeploymentNotifications", func(t *testing.T) {
		t.Skip("Skipping this test since we're not using the deployment notification system")
		dal.DeploymentChanges.Unsubscribe(deploymentChangesCh)
		expectedDeploymentChanges := []DeploymentNotification{
			{Message: optional.Some(Deployment{Language: "go", Module: "test", Schema: &schema.Module{Name: "test"}})},
			{Message: optional.Some(Deployment{Language: "go", Module: "test", MinReplicas: 1, Schema: &schema.Module{Name: "test"}})},
		}
		err = wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, expectedDeploymentChanges, deploymentChanges,
			assert.Exclude[model.DeploymentKey](), assert.Exclude[time.Time](), assert.IgnoreGoStringer())
	})
}

func TestCreateArtefactConflict(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	idch := make(chan sha256.SHA256, 2)

	wg := sync.WaitGroup{}
	wg.Add(2)
	createContent := func() {
		defer wg.Done()
		tx1, err := dal.Begin(ctx)
		assert.NoError(t, err)
		digest, err := tx1.CreateArtefact(ctx, []byte("content"))
		assert.NoError(t, err)
		time.Sleep(time.Second * 2)
		err = tx1.Commit(ctx)
		assert.NoError(t, err)
		idch <- digest
	}

	go createContent()
	go createContent()

	wg.Wait()

	ids := []sha256.SHA256{}

	for range 2 {
		select {
		case id := <-idch:
			ids = append(ids, id)
		case <-time.After(time.Second * 3):
			t.Fatal("Timed out waiting for artefact creation")
		}
	}
	assert.Equal(t, 2, len(ids))
	assert.Equal(t, ids[0], ids[1])
}

func artefactContent(t testing.TB, artefacts []*model.Artefact) [][]byte {
	t.Helper()
	var result [][]byte
	for _, a := range artefacts {
		content, err := io.ReadAll(a.Content)
		assert.NoError(t, err)
		result = append(result, content)
		a.Content = nil
	}
	return result
}

func TestRunnerStateFromProto(t *testing.T) {
	state := ftlv1.RunnerState_RUNNER_NEW
	assert.Equal(t, RunnerStateNew, RunnerStateFromProto(state))
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
	dal, err := New(ctx, conn, encryption.NewBuilder())
	assert.NoError(t, err)

	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = dal.CreateArtefact(ctx, testContent)
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = dal.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
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

func TestVerifyEncryption(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	uri := "fake-kms://CK6YwYkBElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEJy4TIQgfCuwxA3ZZgChp_wYARABGK6YwYkBIAE"

	t.Run("DeleteVerificationColumns", func(t *testing.T) {
		dal, err := New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		// check that there are columns set in encryption_keys
		row, err := dal.db.GetOnlyEncryptionKey(ctx)
		assert.NoError(t, err)
		assert.NotZero(t, row.VerifyTimeline.Ok())
		assert.NotZero(t, row.VerifyAsync.Ok())

		// delete the columns to see if they are recreated
		err = dal.db.UpdateEncryptionVerification(ctx, optional.None[encryption.EncryptedTimelineColumn](), optional.None[encryption.EncryptedAsyncColumn]())
		assert.NoError(t, err)

		dal, err = New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		row, err = dal.db.GetOnlyEncryptionKey(ctx)
		assert.NoError(t, err)
		assert.NotZero(t, row.VerifyTimeline.Ok())
		assert.NotZero(t, row.VerifyAsync.Ok())
	})

	t.Run("DifferentKey", func(t *testing.T) {
		_, err := New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		differentKey := "fake-kms://CJP7ksIKElQKSAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EhIaEJWT3z-xdW23HO7hc9vF3YoYARABGJP7ksIKIAE"
		_, err = New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(differentKey)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decryption failed")
	})

	t.Run("SameKeyButWrongTimelineVerification", func(t *testing.T) {
		dal, err := New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		err = dal.db.UpdateEncryptionVerification(ctx, optional.Some[encryption.EncryptedTimelineColumn]([]byte("123")), optional.None[encryption.EncryptedAsyncColumn]())
		assert.NoError(t, err)
		_, err = New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verification sanity")
		assert.Contains(t, err.Error(), "verify timeline")

		err = dal.db.UpdateEncryptionVerification(ctx, optional.None[encryption.EncryptedTimelineColumn](), optional.Some[encryption.EncryptedAsyncColumn]([]byte("123")))
		assert.NoError(t, err)
		_, err = New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verification sanity")
		assert.Contains(t, err.Error(), "verify async")
	})

	t.Run("SameKeyButEncryptWrongPlainText", func(t *testing.T) {
		result, err := conn.Exec("DELETE FROM encryption_keys")
		assert.NoError(t, err)
		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)
		dal, err := New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.NoError(t, err)

		encrypted := encryption.EncryptedColumn[encryption.TimelineSubKey]{}
		err = dal.encrypt([]byte("123"), &encrypted)
		assert.NoError(t, err)

		err = dal.db.UpdateEncryptionVerification(ctx, optional.Some(encrypted), optional.None[encryption.EncryptedAsyncColumn]())
		assert.NoError(t, err)
		_, err = New(ctx, conn, encryption.NewBuilder().WithKMSURI(optional.Some(uri)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "string does not match")
	})
}
