package leader

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestExistingLeaseForURL(t *testing.T) {
	// Test that if the a lease exists with the current URL, neither a leader or follower is created
	// This can occur when a leader fails to renew a lease, and we then try and coordinate while the db still has the lease active
	// If we create a new follower then we can end up in an infinte loop with the follower calling the leader's service which is really a follower
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	key := leases.SystemKey("leader-test")
	endpoint, _ := url.Parse("http://localhost:1234")
	leaser := leases.NewFakeLeaser()
	_, _, err := leaser.AcquireLease(ctx, key, time.Second*5, optional.Some[any](endpoint.String()))
	if err != nil {
		t.Fatal()
	}
	coordinator := NewCoordinator[interface{}](ctx,
		endpoint,
		key,
		leaser,
		time.Second*10,
		func(ctx context.Context) (interface{}, error) {
			t.Fatal("shouldn't be called")
			return nil, nil
		},
		func(ctx context.Context, endpoint *url.URL) (interface{}, error) {
			t.Fatal("shouldn't be called")
			return nil, nil
		},
	)
	_, err = coordinator.Get()
	assert.Error(t, err)
}

func TestSingleLeader(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	leaser := leases.NewFakeLeaser()
	leaseTTL := time.Second * 5
	leaderFactory := func(ctx context.Context) (string, error) {
		return fmt.Sprintf("leader:%v", time.Now()), nil
	}
	followerFactory := func(ctx context.Context, leaderURL *url.URL) (string, error) {
		fmt.Printf("creating follower with leader url: %s\n", leaderURL.String())
		return fmt.Sprintf("following:%s", leaderURL.String()), nil
	}

	// create coordinators
	coordinators := []*Coordinator[string]{}
	for i := range 5 {
		advertise, err := url.Parse(fmt.Sprintf("http://localhost:%d", i))
		assert.NoError(t, err)
		coordinators = append(coordinators, NewCoordinator[string](ctx,
			advertise,
			leases.SystemKey("leader-test"),
			leaser,
			leaseTTL,
			leaderFactory,
			followerFactory))
	}

	// find leader
	leaderIdx, initialLeaderStr := leaderFromCoordinators(t, coordinators)
	validateAllFollowTheLeader(t, coordinators, leaderIdx)

	// release the lease the leader is using, to simulate the lease not being able to be renewed by the leader
	// a new leader should be elected
	coordinators[leaderIdx].leader.MustGet().lease.Release()
	time.Sleep(leaseTTL + time.Millisecond*500)

	leaderIdx, finalLeaderStr := leaderFromCoordinators(t, coordinators)
	assert.NotEqual(t, finalLeaderStr, initialLeaderStr, "leader should have been changed when the lease broke")
	validateAllFollowTheLeader(t, coordinators, leaderIdx)
}

func leaderFromCoordinators(t *testing.T, coordinators []*Coordinator[string]) (idx int, leaderStr string) {
	t.Helper()

	leaderIdx := -1
	for i := range 5 {
		result, err := coordinators[i].Get()
		assert.NoError(t, err)
		if strings.HasPrefix(result, "leader:") {
			leaderIdx = i
			leaderStr = result
		}
	}
	assert.NotEqual(t, -1, leaderIdx)
	return idx, leaderStr
}

func validateAllFollowTheLeader(t *testing.T, coordinators []*Coordinator[string], leaderIdx int) {
	t.Helper()

	for i := range 5 {
		if leaderIdx == i {
			// known leader
			continue
		}
		result, err := coordinators[i].Get()
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("following:%s", coordinators[leaderIdx].advertise), result)
	}
}
