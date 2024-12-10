package state_test

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/state"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/model"
)

func TestRunnerState(t *testing.T) {
	cs := state.NewInMemoryState()
	view := cs.View()
	assert.Equal(t, 0, len(view.Runners()))
	key := model.NewLocalRunnerKey(1)
	create := time.Now()
	endpoint := "http://localhost:8080"
	module := "test"
	deploymentKey := model.NewDeploymentKey(module)
	err := cs.Publish(&state.RunnerRegisteredEvent{
		Key:        key,
		Time:       create,
		Endpoint:   endpoint,
		Module:     module,
		Deployment: deploymentKey,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 1, len(view.Runners()))
	assert.Equal(t, key, view.Runners()[0].Key)
	assert.Equal(t, create, view.Runners()[0].Create)
	assert.Equal(t, create, view.Runners()[0].LastSeen)
	assert.Equal(t, endpoint, view.Runners()[0].Endpoint)
	assert.Equal(t, module, view.Runners()[0].Module)
	assert.Equal(t, deploymentKey, view.Runners()[0].Deployment)
	seen := time.Now()
	err = cs.Publish(&state.RunnerRegisteredEvent{
		Key:        key,
		Time:       seen,
		Endpoint:   endpoint,
		Module:     module,
		Deployment: deploymentKey,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, seen, view.Runners()[0].LastSeen)

	err = cs.Publish(&state.RunnerDeletedEvent{
		Key: key,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 0, len(view.Runners()))

}

func TestDeploymentState(t *testing.T) {
	cs := state.NewInMemoryState()
	view := cs.View()
	assert.Equal(t, 0, len(view.Deployments()))

	deploymentKey := model.NewDeploymentKey("test-deployment")
	create := time.Now()
	err := cs.Publish(&state.DeploymentCreatedEvent{
		Key:       deploymentKey,
		CreatedAt: create,
		Schema:    &schemapb.Module{Name: "test"},
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 1, len(view.Deployments()))
	assert.Equal(t, deploymentKey, view.Deployments()[deploymentKey.String()].Key)
	assert.Equal(t, create, view.Deployments()[deploymentKey.String()].CreatedAt)

	activate := time.Now()
	err = cs.Publish(&state.DeploymentActivatedEvent{
		Key:         deploymentKey,
		ActivatedAt: activate,
		MinReplicas: 1,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 1, view.Deployments()[deploymentKey.String()].MinReplicas)
	assert.Equal(t, activate, view.Deployments()[deploymentKey.String()].ActivatedAt.MustGet())

	err = cs.Publish(&state.DeploymentDeactivatedEvent{
		Key: deploymentKey,
	})
	assert.NoError(t, err)
	view = cs.View()

	assert.Equal(t, 0, view.Deployments()[deploymentKey.String()].MinReplicas)
}
