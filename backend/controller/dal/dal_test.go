package dal

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types"
)

//nolint:maintidx
func TestDAL(t *testing.T) {
	logger := log.Configure(os.Stderr, log.Config{Level: log.Debug})
	ctx := log.ContextWithLogger(context.Background(), logger)
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)
	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSHA = sha256.Sum(testContent)

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
	var deploymentName model.DeploymentName
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentName, err = dal.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil)
		assert.NoError(t, err)
	})

	deployment := &model.Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Name:     deploymentName,
		Artefacts: []*model.Artefact{
			{Path: "dir/filename",
				Executable: true,
				Digest:     testSHA,
				Content:    io.NopCloser(bytes.NewReader(testContent))},
		},
	}
	expectedContent := artefactContent(t, deployment.Artefacts)

	t.Run("GetDeployment", func(t *testing.T) {
		actual, err := dal.GetDeployment(ctx, deploymentName)
		assert.NoError(t, err)
		actualContent := artefactContent(t, actual.Artefacts)
		assert.Equal(t, expectedContent, actualContent)
		assert.Equal(t, deployment, actual)
	})

	t.Run("GetMissingDeployment", func(t *testing.T) {
		_, err := dal.GetDeployment(ctx, model.NewDeploymentName("test"))
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("GetMissingArtefacts", func(t *testing.T) {
		misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
		missing, err := dal.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
		assert.NoError(t, err)
		assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
	})

	runnerID := model.NewRunnerKey()
	labels := map[string]any{"languages": []any{"go"}}

	t.Run("RegisterRunner", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:      runnerID,
			Labels:   labels,
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		})
		assert.NoError(t, err)
	})

	t.Run("RegisterRunnerFailsOnDuplicate", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:      model.NewRunnerKey(),
			Labels:   labels,
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		})
		assert.Error(t, err)
		assert.IsError(t, err, ErrConflict)
	})

	t.Run("GetIdleRunnersForLanguage", func(t *testing.T) {
		expectedRunner := Runner{
			Key:      runnerID,
			Labels:   labels,
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		}
		runners, err := dal.GetIdleRunners(ctx, 10, labels)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{expectedRunner}, runners)
	})

	expectedRunner := Runner{
		Key:        runnerID,
		Labels:     labels,
		Endpoint:   "http://localhost:8080",
		State:      RunnerStateReserved,
		Deployment: types.Some(deploymentName),
	}

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{}, reconcile)
	})

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := dal.SetDeploymentReplicas(ctx, deploymentName, 1)
		assert.NoError(t, err)
	})

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{{
			Deployment:       deploymentName,
			Module:           deployment.Module,
			Language:         deployment.Language,
			AssignedReplicas: 0,
			RequiredReplicas: 1,
		}}, reconcile)
	})

	t.Run("ReserveRunnerForInvalidDeployment", func(t *testing.T) {
		_, err := dal.ReserveRunnerForDeployment(ctx, model.NewDeploymentName("test"), time.Second, labels)
		assert.Error(t, err)
		assert.IsError(t, err, ErrNotFound)
		assert.EqualError(t, err, "deployment: not found")
	})

	t.Run("ReserveRunnerForDeployment", func(t *testing.T) {
		claim, err := dal.ReserveRunnerForDeployment(ctx, deploymentName, time.Millisecond*100, labels)
		assert.NoError(t, err)
		err = claim.Commit(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, claim.Runner())
	})

	t.Run("ExpireRunnerClaims", func(t *testing.T) {
		time.Sleep(time.Millisecond * 500)
		count, err := dal.ExpireRunnerClaims(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		runners, err := dal.GetIdleRunners(ctx, 10, labels)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(runners))
	})

	t.Run("ReserveRunnerForDeploymentFailsOnInvalidDeployment", func(t *testing.T) {
		_, err = dal.ReserveRunnerForDeployment(ctx, model.NewDeploymentName("test"), time.Second, labels)
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("UpdateRunnerAssigned", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(deploymentName),
		})
		assert.NoError(t, err)
	})

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{}, reconcile)
	})

	t.Run("GetRunnersForDeployment", func(t *testing.T) {
		runners, err := dal.GetRunnersForDeployment(ctx, deploymentName)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(deploymentName),
		}}, runners)
	})

	var requestName model.RequestName
	t.Run("CreateIngressRequest", func(t *testing.T) {
		requestName, err = dal.CreateIngressRequest(ctx, "GET /test", "127.0.0.1:1234")
		assert.NoError(t, err)
	})

	callEvent := &CallEvent{
		Time:           time.Now().Round(time.Millisecond),
		DeploymentName: deploymentName,
		RequestName:    types.Some(requestName),
		Request:        []byte("{}"),
		Response:       []byte(`{"time": "now"}`),
		DestVerb:       schema.VerbRef{Module: "time", Name: "time"},
	}
	t.Run("InsertCallEvent", func(t *testing.T) {
		err = dal.InsertCallEvent(ctx, callEvent)
		assert.NoError(t, err)
	})

	logEvent := &LogEvent{
		Time:           time.Now().Round(time.Millisecond),
		DeploymentName: deploymentName,
		RequestName:    types.Some(requestName),
		Level:          int32(log.Warn),
		Attributes:     map[string]string{"attr": "value"},
		Message:        "A log entry",
	}
	t.Run("InsertLogEntry", func(t *testing.T) {
		err = dal.InsertLogEvent(ctx, logEvent)
		assert.NoError(t, err)
	})

	expectedDeploymentUpdatedEvent := &DeploymentUpdatedEvent{
		DeploymentName: deploymentName,
		MinReplicas:    1,
	}

	t.Run("QueryEvents", func(t *testing.T) {
		t.Run("Limit", func(t *testing.T) {
			events, err := dal.QueryEvents(ctx, 1)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(events))
		})

		t.Run("NoFilters", func(t *testing.T) {
			events, err := dal.QueryEvents(ctx, 1000)
			assert.NoError(t, err)
			assertEventsEqual(t, []Event{expectedDeploymentUpdatedEvent, callEvent, logEvent}, events)
		})

		t.Run("ByDeployment", func(t *testing.T) {
			events, err := dal.QueryEvents(ctx, 1000, FilterDeployments(deploymentName))
			assert.NoError(t, err)
			assertEventsEqual(t, []Event{expectedDeploymentUpdatedEvent, callEvent, logEvent}, events)
		})

		t.Run("ByCall", func(t *testing.T) {
			events, err := dal.QueryEvents(ctx, 1000, FilterTypes(EventTypeCall), FilterCall(types.None[string](), "time", types.None[string]()))
			assert.NoError(t, err)
			assertEventsEqual(t, []Event{callEvent}, events)
		})

		t.Run("ByLogLevel", func(t *testing.T) {
			events, err := dal.QueryEvents(ctx, 1000, FilterTypes(EventTypeLog), FilterLogLevel(log.Trace))
			assert.NoError(t, err)
			assertEventsEqual(t, []Event{logEvent}, events)
		})
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		routes, err := dal.GetRoutingTable(ctx, []string{deployment.Module})
		assert.NoError(t, err)
		assert.Equal(t, []Route{{
			Module:     "test",
			Runner:     expectedRunner.Key,
			Deployment: deploymentName,
			Endpoint:   expectedRunner.Endpoint,
		}}, routes[deployment.Module])
	})

	t.Run("UpdateRunnerInvalidDeployment", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Labels:     labels,
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(model.NewDeploymentName("test")),
		})
		assert.Error(t, err)
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("ReleaseRunnerReservation", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:      runnerID,
			Labels:   labels,
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		})
		assert.NoError(t, err)
	})

	t.Run("ReserveRunnerForDeploymentAfterRelease", func(t *testing.T) {
		claim, err := dal.ReserveRunnerForDeployment(ctx, deploymentName, time.Second, labels)
		assert.NoError(t, err)
		err = claim.Commit(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, claim.Runner())
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		_, err := dal.GetRoutingTable(ctx, []string{deployment.Module})
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("DeregisterRunner", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, runnerID)
		assert.NoError(t, err)
	})

	t.Run("DeregisterRunnerFailsOnMissing", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, model.NewRunnerKey())
		assert.IsError(t, err, ErrNotFound)
	})
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
	state := ftlv1.RunnerState_RUNNER_IDLE
	assert.Equal(t, RunnerStateIdle, RunnerStateFromProto(state))
}

func TestControllerStateFromProto(t *testing.T) {
	state := ftlv1.ControllerState_CONTROLLER_LIVE
	assert.Equal(t, ControllerStateLive, ControllerStateFromProto(state))
}

func normaliseEvents(events []Event) []Event {
	for i := 0; i < len(events); i++ {
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

func assertEventsEqual(t *testing.T, expected, actual []Event) {
	t.Helper()
	assert.Equal(t, normaliseEvents(expected), normaliseEvents(actual))
}
