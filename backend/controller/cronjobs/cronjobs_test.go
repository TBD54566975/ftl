package cronjobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"
	xslices "golang.org/x/exp/slices"

	db "github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

func TestServiceWithMockDal(t *testing.T) {
	t.Skip("TODO: sometimes blocks on CI. Discussion in issue #1368")
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	clk := clock.NewMock()
	clk.Add(time.Second) // half way between cron job executions

	mockDal := &mockDAL{
		clock:           clk,
		lock:            sync.Mutex{},
		attemptCountMap: map[string]int{},
	}
	conn := sqltest.OpenForTesting(ctx, t)
	parentDAL, err := db.New(ctx, conn, encryption.NewForKey([]byte{}))
	assert.NoError(t, err)

	testServiceWithDal(ctx, t, mockDal, parentDAL, clk)
}

func TestHashRing(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	// This test uses multiple mock clocks to progress time for each controller individually
	// This allows us to compare attempts for each cron job and know which controller attempted it
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	mockDal := &mockDAL{
		clock:           clock.NewMock(),
		lock:            sync.Mutex{},
		attemptCountMap: map[string]int{},
	}
	moduleName := "initial"
	jobsToCreate := newJobs(t, moduleName, "*/10 * * * * * *", mockDal.clock, 100)

	deploymentKey, err := mockDal.CreateDeployment(ctx, "go", &schema.Module{
		Name: moduleName,
	}, []db.DeploymentArtefact{}, []db.IngressRoutingEntry{}, jobsToCreate)
	assert.NoError(t, err)

	err = mockDal.ReplaceDeployment(ctx, deploymentKey, 1)
	assert.NoError(t, err)

	controllers := newControllers(ctx, 20, mockDal, func() clock.Clock { return clock.NewMock() }, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
		return &connect.Response[ftlv1.CallResponse]{}, nil
	})

	// This should give time for each controller to start watching its own mock clock
	// If we don;t wait here, we might hit a race condition outlined in issue #1368
	time.Sleep(time.Millisecond * 100)

	// progress time for each controller one at a time, noting which verbs got attempted each time
	// to build a map of verb to controller keys
	controllersForVerbs := map[string][]model.ControllerKey{}
	for _, c := range controllers {
		mockDal.lock.Lock()
		beforeAttemptCount := map[string]int{}
		for k, v := range mockDal.attemptCountMap {
			beforeAttemptCount[k] = v
		}
		mockDal.lock.Unlock()

		c.mockClock.Add(time.Second * 15)
		time.Sleep(time.Millisecond * 100)

		mockDal.lock.Lock()
		for k, v := range mockDal.attemptCountMap {
			if beforeAttemptCount[k] == v {
				continue
			}
			controllersForVerbs[k] = append(controllersForVerbs[k], c.key)
		}
		mockDal.lock.Unlock()
	}

	// Check if each job has the same key list
	// Theoretically this is is possible for all jobs to have the same assigned controllers, but with 100 jobs and 20 controllers, this is unlikely
	keys := []string{}
	hasFoundNonMatchingKeys := false
	for v, k := range controllersForVerbs {
		assert.Equal(t, len(k), 2, "expected verb %s to be attempted by 2 controllers", v)

		kStrs := slices.Map(k, func(k model.ControllerKey) string { return k.String() })
		xslices.Sort(kStrs)
		if len(keys) == 0 {
			keys = kStrs
			continue
		}

		if hasFoundNonMatchingKeys == false {
			for keyIdx, keyStr := range kStrs {
				if keys[keyIdx] != keyStr {
					hasFoundNonMatchingKeys = true
				}
			}
		}
	}
	assert.True(t, hasFoundNonMatchingKeys, "expected at least one verb to have different controllers assigned")
}
