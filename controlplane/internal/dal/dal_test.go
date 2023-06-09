package dal

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types"
	"github.com/oklog/ulid/v2"

	"github.com/TBD54566975/ftl/controlplane/internal/sql/sqltest"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/schema"
)

func TestDAL(t *testing.T) {
	var testContent = bytes.Repeat([]byte("sometestcontentthatislongerthanthereadbuffer"), 100)
	var testSHA = sha256.Sum(testContent)

	conn := sqltest.OpenForTesting(t)
	dal := New(conn)
	assert.NotZero(t, dal)
	var err error

	ctx := context.Background()

	t.Run("CreateModule", func(t *testing.T) {
		err = dal.CreateModule(ctx, "go", "test")
		assert.NoError(t, err)
	})

	var testSha sha256.SHA256

	t.Run("CreateArtefact", func(t *testing.T) {
		testSha, err = dal.CreateArtefact(ctx, testContent)
		assert.NoError(t, err)
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey ulid.ULID
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = dal.CreateDeployment(ctx, "go", module, []DeploymentArtefact{{
			Digest:     testSha,
			Executable: true,
			Path:       "dir/filename",
		}})
		assert.NoError(t, err)
	})

	deployment := &Deployment{
		Module:   "test",
		Language: "go",
		Schema:   module,
		Key:      deploymentKey,
		Artefacts: []*Artefact{
			{Path: "dir/filename",
				Executable: true,
				Digest:     testSHA,
				Content:    bytes.NewReader(testContent)},
		},
	}
	expectedContent := artefactContent(t, deployment.Artefacts)

	t.Run("GetLatestDeployment", func(t *testing.T) {
		actual, err := dal.GetLatestDeployment(ctx, deployment.Module)
		assert.NoError(t, err)
		actualContent := artefactContent(t, actual.Artefacts)
		assert.Equal(t, expectedContent, actualContent)
		assert.Equal(t, deployment, actual)
	})

	t.Run("GetDeployment", func(t *testing.T) {
		actual, err := dal.GetDeployment(ctx, deploymentKey)
		assert.NoError(t, err)
		actualContent := artefactContent(t, actual.Artefacts)
		assert.Equal(t, expectedContent, actualContent)
		assert.Equal(t, deployment, actual)
	})

	t.Run("GetMissingDeployment", func(t *testing.T) {
		_, err := dal.GetDeployment(ctx, ulid.Make())
		assert.EqualError(t, err, ErrNotFound.Error())
	})

	t.Run("GetMissingArtefacts", func(t *testing.T) {
		misshingSHA := sha256.MustParseSHA256("fae7e4cbdca7167bbea4098c05d596f50bbb18062b61c1dfca3705b4a6c2888c")
		missing, err := dal.GetMissingArtefacts(ctx, []sha256.SHA256{testSHA, misshingSHA})
		assert.NoError(t, err)
		assert.Equal(t, []sha256.SHA256{misshingSHA}, missing)
	})

	runnerID := ulid.Make()
	t.Run("RegisterRunner", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:      runnerID,
			Language: "go",
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		})
		assert.NoError(t, err)
	})

	t.Run("RegisterRunnerFailsOnDuplicate", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:      ulid.Make(),
			Language: "go",
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		})
		assert.Error(t, err)
		assert.IsError(t, err, ErrConflict)
	})

	t.Run("GetIdleRunnersForLanguage", func(t *testing.T) {
		expectedRunner := Runner{
			Key:      runnerID,
			Language: "go",
			Endpoint: "http://localhost:8080",
			State:    RunnerStateIdle,
		}
		runners, err := dal.GetIdleRunnersForLanguage(ctx, "go", 10)
		assert.NoError(t, err)
		assert.Equal(t, []Runner{expectedRunner}, runners)
	})

	expectedRunner := Runner{
		Key:      runnerID,
		Language: "go",
		Endpoint: "http://localhost:8080",
		State:    RunnerStateReserved,
	}

	t.Run("ReserveRunnerForInvalidDeployment", func(t *testing.T) {
		_, err := dal.ReserveRunnerForDeployment(ctx, "go", ulid.Make())
		assert.IsError(t, err, ErrInvalidReference)
		assert.EqualError(t, err, "deployment: invalid reference")
	})

	t.Run("ReserveRunnerForDeployment", func(t *testing.T) {
		actualRunner, err := dal.ReserveRunnerForDeployment(ctx, "go", deploymentKey)
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, actualRunner)
	})

	t.Run("ReserveRunnerForDeploymentFailsOnDuplicate", func(t *testing.T) {
		_, err = dal.ReserveRunnerForDeployment(ctx, "go", ulid.Make())
		assert.IsError(t, err, ErrNotFound)
		assert.EqualError(t, err, `no idle runners for language "go": not found`)
	})

	t.Run("UpdateRunnerAssigned", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Language:   "go",
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(deploymentKey),
		})
		assert.NoError(t, err)
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		routes, err := dal.GetRoutingTable(ctx, deployment.Module)
		assert.NoError(t, err)
		assert.Equal(t, []string{expectedRunner.Endpoint}, routes)
	})

	t.Run("UpdateRunnerInvalidDeployment", func(t *testing.T) {
		err := dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Language:   "go",
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateAssigned,
			Deployment: types.Some(ulid.Make()),
		})
		assert.Error(t, err)
	})

	t.Run("ReleaseRunnerReservation", func(t *testing.T) {
		err = dal.UpsertRunner(ctx, Runner{
			Key:        runnerID,
			Language:   "go",
			Endpoint:   "http://localhost:8080",
			State:      RunnerStateIdle,
			Deployment: types.Some(deploymentKey),
		})
		assert.NoError(t, err)
	})

	t.Run("ReserveRunnerForDeploymentAfterRelease", func(t *testing.T) {
		actualRunner, err := dal.ReserveRunnerForDeployment(ctx, "go", deploymentKey)
		assert.NoError(t, err)
		assert.Equal(t, expectedRunner, actualRunner)
	})

	t.Run("GetRoutingTable", func(t *testing.T) {
		_, err := dal.GetRoutingTable(ctx, deployment.Module)
		assert.EqualError(t, err, "not found")
	})

	t.Run("DeregisterRunner", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, runnerID)
		assert.NoError(t, err)
	})

	t.Run("DeregisterRunnerFailsOnMissing", func(t *testing.T) {
		err = dal.DeregisterRunner(ctx, runnerID)
		assert.EqualError(t, err, "not found")
	})
}

func artefactContent(t testing.TB, artefacts []*Artefact) [][]byte {
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
