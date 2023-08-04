package dal

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types"

	log2 "github.com/TBD54566975/ftl/backend/common/log"
	model2 "github.com/TBD54566975/ftl/backend/common/model"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/controller/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

func TestDAL(t *testing.T) {
	logger := log2.Configure(os.Stderr, log2.Config{Level: log2.Debug})
	ctx := log2.ContextWithLogger(context.Background(), logger)
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
	var deploymentKey model2.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = dal.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}}, nil)
		assert.NoError(t, err)
	})

	deployment := &model2.Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Key:      deploymentKey,
		Artefacts: []*model2.Artefact{
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
		_, err := dal.GetDeployment(ctx, model2.NewDeploymentKey())
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("GetMissingArtefacts", func(t *testing.T) {
		misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
		missing, err := dal.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
		assert.NoError(t, err)
		assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
	})

	runnerID := model2.NewRunnerKey()
	t.Run("RegisterRunner", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:       runnerID,
			Languages: []string{"go"},
			Endpoint:  "http://localhost:8080",
			State:     RunnerStateIdle,
		})
		assert.NoError(t, err)
	})

	t.Run("RegisterRunnerFailsOnDuplicate", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:       model2.NewRunnerKey(),
			Languages: []string{"go"},
			Endpoint:  "http://localhost:8080",
			State:     RunnerStateIdle,
		})
		assert.Error(t, err)
		assert.IsError(t, err, ErrConflict)
	})

	t.Run("GetIdleRunnersForLanguage", func(t *testing.T) {
		expectedRunner := Runner{
			Key:       runnerID,
			Languages: []string{"go"},
			Endpoint:  "http://localhost:8080",
			State:     RunnerStateIdle,
		}
		runners, err := dal.GetIdleRunnersForLanguage(ctx, "go", 10)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{expectedRunner}, runners)
	})

	expectedRunner := Runner{
		Key:        runnerID,
		Languages:  []string{"go"},
		Endpoint:   "http://localhost:8080",
		State:      RunnerStateReserved,
		Deployment: types.Some(deploymentKey),
	}

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{}, reconcile)
	})

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := dal.SetDeploymentReplicas(ctx, deploymentKey, 1)
		assert.NoError(t, err)
	})

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{{
			Deployment:       deploymentKey,
			Module:           deployment.Module,
			Language:         deployment.Language,
			AssignedReplicas: 0,
			RequiredReplicas: 1,
		}}, reconcile)
	})

	t.Run("ReserveRunnerForInvalidDeployment", func(t *testing.T) {
		_, err := dal.ReserveRunnerForDeployment(ctx, "go", model2.NewDeploymentKey(), time.Second)
		assert.Error(t, err)
		assert.IsError(t, err, ErrNotFound)
		assert.EqualError(t, err, "deployment: not found")
	})

	t.Run("ReserveRunnerForDeployment", func(t *testing.T) {
		claim, err := dal.ReserveRunnerForDeployment(ctx, "go", deploymentKey, time.Millisecond*100)
		assert.NoError(t, err)
		err = claim.Commit(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, claim.Runner())
	})

	t.Run("ExpireRunnerClaims", func(t *testing.T) {
		time.Sleep(time.Millisecond * 200)
		count, err := dal.ExpireRunnerClaims(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		runners, err := dal.GetIdleRunnersForLanguage(ctx, "go", 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(runners))
	})

	t.Run("ReserveRunnerForDeploymentFailsOnInvalidDeployment", func(t *testing.T) {
		_, err = dal.ReserveRunnerForDeployment(ctx, "go", model2.NewDeploymentKey(), time.Second)
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("UpdateRunnerAssigned", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Languages:  []string{"go"},
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(deploymentKey),
		})
		assert.NoError(t, err)
	})

	t.Run("GetDeploymentsNeedingReconciliation", func(t *testing.T) {
		reconcile, err := dal.GetDeploymentsNeedingReconciliation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []Reconciliation{}, reconcile)
	})

	t.Run("GetRunnersForDeployment", func(t *testing.T) {
		runners, err := dal.GetRunnersForDeployment(ctx, deploymentKey)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{{
			Key:        runnerID,
			Languages:  []string{"go"},
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(deploymentKey),
		}}, runners)
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		routes, err := dal.GetRoutingTable(ctx, deployment.Module)
		assert.NoError(t, err)
		assert.Equal(t, []Route{{
			Runner:   expectedRunner.Key,
			Endpoint: expectedRunner.Endpoint,
		}}, routes)
	})

	t.Run("UpdateRunnerInvalidDeployment", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Languages:  []string{"go"},
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(model2.NewDeploymentKey()),
		})
		assert.Error(t, err)
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("ReleaseRunnerReservation", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:       runnerID,
			Languages: []string{"go"},
			Endpoint:  "http://localhost:8080",
			State:     RunnerStateIdle,
		})
		assert.NoError(t, err)
	})

	t.Run("ReserveRunnerForDeploymentAfterRelease", func(t *testing.T) {
		claim, err := dal.ReserveRunnerForDeployment(ctx, "go", deploymentKey, time.Second)
		assert.NoError(t, err)
		err = claim.Commit(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, claim.Runner())
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		_, err := dal.GetRoutingTable(ctx, deployment.Module)
		assert.IsError(t, err, ErrNotFound)
	})

	t.Run("DeregisterRunner", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, runnerID)
		assert.NoError(t, err)
	})

	t.Run("DeregisterRunnerFailsOnMissing", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, model2.NewRunnerKey())
		assert.IsError(t, err, ErrNotFound)
	})
}

func artefactContent(t testing.TB, artefacts []*model2.Artefact) [][]byte {
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
