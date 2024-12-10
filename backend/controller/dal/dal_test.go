package dal

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	"github.com/TBD54566975/ftl/backend/timeline"

	dalmodel "github.com/TBD54566975/ftl/backend/controller/dal/model"
	"github.com/TBD54566975/ftl/backend/controller/pubsub"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
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
	conn := sqltest.OpenForTesting(ctx, t)

	pubSub := pubsub.New(ctx, conn, optional.None[pubsub.AsyncCallListener]())
	dal := New(ctx, conn, pubSub, artefacts.NewForTesting())

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

	module := &schema.Module{Name: "test"}
	var deploymentKey model.DeploymentKey
	t.Run("CreateDeployment", func(t *testing.T) {
		deploymentKey, err = dal.CreateDeployment(ctx, "go", module)
		assert.NoError(t, err)
	})

	t.Run("SetDeploymentReplicas", func(t *testing.T) {
		err := dal.SetDeploymentReplicas(ctx, deploymentKey, 1)
		assert.NoError(t, err)
	})

	t.Run("VerifyDeploymentNotifications", func(t *testing.T) {
		t.Skip("Skipping this test since we're not using the deployment notification system")
		dal.DeploymentChanges.Unsubscribe(deploymentChangesCh)
		expectedDeploymentChanges := []DeploymentNotification{
			{Message: optional.Some(dalmodel.Deployment{Language: "go", Module: "test", Schema: &schema.Module{Name: "test"}})},
			{Message: optional.Some(dalmodel.Deployment{Language: "go", Module: "test", MinReplicas: 1, Schema: &schema.Module{Name: "test"}})},
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

	pubSub := pubsub.New(ctx, conn, optional.None[pubsub.AsyncCallListener]())

	dal := New(ctx, conn, pubSub, artefacts.NewForTesting())

	idch := make(chan sha256.SHA256, 2)

	wg := sync.WaitGroup{}
	wg.Add(2)
	createContent := func() {
		defer wg.Done()
		tx1, err := dal.Begin(ctx)
		assert.NoError(t, err)
		digest, err := dal.registry.Upload(ctx, artefacts.Artefact{Content: []byte("content")})
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
