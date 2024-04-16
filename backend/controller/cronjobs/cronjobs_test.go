//go:build !integration

package cronjobs

import (
	"context"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"
	xslices "golang.org/x/exp/slices"
)

func TestService(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	clk := clock.NewMock()
	mockDal := &mockDAL{
		clock:           clk,
		lock:            sync.Mutex{},
		attemptCountMap: map[string]int{},
	}

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	moduleName := "initial"
	deploymentKey := model.NewDeploymentKey(moduleName)
	for _, job := range newJobs(t, moduleName, "*/10 * * * * * *", clk, 20) {
		mockDal.createCronJob(deploymentKey, moduleName, job.Verb.Name, job.Schedule, job.StartTime, job.NextExecution)
	}

	_ = newControllers(ctx, 5, mockDal, func() clock.Clock { return clk }, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
		verbRef := schema.RefFromProto(r.Msg.Verb)

		verbCallCountLock.Lock()
		verbCallCount[verbRef.Name]++
		verbCallCountLock.Unlock()

		return &connect.Response[ftlv1.CallResponse]{}, nil
	})

	clk.Add(time.Second * 5)
	time.Sleep(time.Millisecond * 100)
	for range 3 {
		clk.Add(time.Second * 10)
		time.Sleep(time.Millisecond * 100)
	}

	for _, j := range mockDal.jobs {
		count := verbCallCount[j.Verb.Name]
		assert.Equal(t, count, 3, "expected verb %s to be called 3 times", j.Verb.Name)
	}
}

func TestHashRing(t *testing.T) {
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

	verbCallCount := map[string]int{}
	verbCallCountLock := sync.Mutex{}

	moduleName := "initial"
	deploymentKey := model.NewDeploymentKey(moduleName)
	for _, job := range newJobs(t, moduleName, "*/10 * * * * * *", mockDal.clock, 100) {
		mockDal.createCronJob(deploymentKey, moduleName, job.Verb.Name, job.Schedule, job.StartTime, job.NextExecution)
	}

	controllers := newControllers(ctx, 20, mockDal, func() clock.Clock { return clock.NewMock() }, func(ctx context.Context, r *connect.Request[ftlv1.CallRequest], o optional.Option[model.RequestKey], s string) (*connect.Response[ftlv1.CallResponse], error) {
		verbRef := schema.RefFromProto(r.Msg.Verb)

		verbCallCountLock.Lock()
		verbCallCount[verbRef.Name]++
		verbCallCountLock.Unlock()

		return &connect.Response[ftlv1.CallResponse]{}, nil
	})

	// progress time for each controller one at a time, noting which verbs got attempted each time
	// to build a map of verb to controller keys
	controllersForVerbs := map[string][]model.ControllerKey{}
	for _, c := range controllers {
		beforeAttemptCount := map[string]int{}
		for k, v := range mockDal.attemptCountMap {
			beforeAttemptCount[k] = v
		}

		c.mockClock.Add(time.Second * 15)
		time.Sleep(time.Millisecond * 100)

		for k, v := range mockDal.attemptCountMap {
			if beforeAttemptCount[k] == v {
				continue
			}
			controllersForVerbs[k] = append(controllersForVerbs[k], c.key)
		}
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
