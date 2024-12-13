package state_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/log"
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
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	err := cs.Publish(ctx, &state.RunnerRegisteredEvent{
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
	err = cs.Publish(ctx, &state.RunnerRegisteredEvent{
		Key:        key,
		Time:       seen,
		Endpoint:   endpoint,
		Module:     module,
		Deployment: deploymentKey,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, seen, view.Runners()[0].LastSeen)

	err = cs.Publish(ctx, &state.RunnerDeletedEvent{
		Key: key,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 0, len(view.Runners()))

}

func TestDeploymentState(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cs := state.NewInMemoryState()
	view := cs.View()
	assert.Equal(t, 0, len(view.GetDeployments()))

	deploymentKey := model.NewDeploymentKey("test-deployment")
	create := time.Now()
	err := cs.Publish(ctx, &state.DeploymentCreatedEvent{
		Key:       deploymentKey,
		CreatedAt: create,
		Schema:    &schema.Module{Name: "test"},
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 1, len(view.GetDeployments()))
	assert.Equal(t, deploymentKey, view.GetDeployments()[deploymentKey.String()].Key)
	assert.Equal(t, create, view.GetDeployments()[deploymentKey.String()].CreatedAt)

	activate := time.Now()
	err = cs.Publish(ctx, &state.DeploymentActivatedEvent{
		Key:         deploymentKey,
		ActivatedAt: activate,
		MinReplicas: 1,
	})
	assert.NoError(t, err)
	view = cs.View()
	assert.Equal(t, 1, view.GetDeployments()[deploymentKey.String()].MinReplicas)
	assert.Equal(t, activate, view.GetDeployments()[deploymentKey.String()].ActivatedAt.MustGet())

	err = cs.Publish(ctx, &state.DeploymentDeactivatedEvent{
		Key: deploymentKey,
	})
	assert.NoError(t, err)
	view = cs.View()

	assert.Equal(t, 0, view.GetDeployments()[deploymentKey.String()].MinReplicas)
}
