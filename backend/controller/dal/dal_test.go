package dal

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/controller/state"
	"github.com/TBD54566975/ftl/backend/timeline"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/sha256"
)

func TestDAL(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	timelineEndpoint, err := url.Parse("http://localhost:8080")
	assert.NoError(t, err)
	ctx = timeline.ContextWithClient(ctx, timeline.NewClient(ctx, timelineEndpoint))

	dal := New(artefacts.NewForTesting(), state.NewInMemoryState())

	deploymentChangesCh := dal.DeploymentChanges.Subscribe(nil)
	deploymentChanges := []DeploymentNotification{}
	wg := errgroup.Group{}
	wg.Go(func() error {
		for change := range deploymentChangesCh {
			deploymentChanges = append(deploymentChanges, change)
		}
		return nil
	})

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey = model.NewDeploymentKey(module.Name)
		err = dal.state.Publish(ctx, &state.DeploymentCreatedEvent{
			Key:       deploymentKey,
			CreatedAt: time.Now(),
			Module:    module.Name,
			Schema:    &schema.Module{Name: module.Name},
		})
		assert.NoError(t, err)
	})

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := dal.SetDeploymentReplicas(ctx, deploymentKey, 1)
		assert.NoError(t, err)
	})
}

func TestCreateArtefactConflict(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	dal := New(artefacts.NewForTesting(), state.NewInMemoryState())

	idch := make(chan sha256.SHA256, 2)

	wg := sync.WaitGroup{}

	wg.Add(2)
	createContent := func() {
		defer wg.Done()
		digest, err := dal.registry.Upload(ctx, artefacts.Artefact{Content: []byte("content")})
		assert.NoError(t, err)
		time.Sleep(time.Second * 2)
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
