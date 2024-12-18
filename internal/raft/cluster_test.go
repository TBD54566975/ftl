package raft_test

import (
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/raft"
	"golang.org/x/sync/errgroup"
)

type IntEvent int64

func (e *IntEvent) UnmarshalBinary(data []byte) error {
	*e = IntEvent(binary.BigEndian.Uint64(data))
	return nil
}

func (e IntEvent) MarshalBinary() ([]byte, error) {
	return binary.BigEndian.AppendUint64([]byte{}, uint64(e)), nil
}

type IntStateMachine struct {
	sum int64
}

var _ raft.StateMachine[int64, int64, IntEvent, *IntEvent] = &IntStateMachine{}

func (s *IntStateMachine) Update(event IntEvent) error {
	s.sum += int64(event)
	return nil
}

func (s *IntStateMachine) Lookup(key int64) (int64, error) { return s.sum, nil }
func (s *IntStateMachine) Recover(reader io.Reader) error  { return nil }
func (s *IntStateMachine) Save(writer io.Writer) error     { return nil }
func (s *IntStateMachine) Close() error                    { return nil }

func TestCluster(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()

	members := []string{"localhost:51001", "localhost:51002"}

	cluster1 := testCluster(t, members, 1, members[0])
	shard1_1 := raft.AddShard(ctx, cluster1, 1, &IntStateMachine{})
	shard1_2 := raft.AddShard(ctx, cluster1, 2, &IntStateMachine{})

	cluster2 := testCluster(t, members, 2, members[1])
	shard2_1 := raft.AddShard(ctx, cluster2, 1, &IntStateMachine{})
	shard2_2 := raft.AddShard(ctx, cluster2, 2, &IntStateMachine{})

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	assert.NoError(t, wg.Wait())
	defer cluster1.Stop()
	defer cluster2.Stop()

	assert.NoError(t, shard1_1.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_1.Propose(ctx, IntEvent(2)))

	assert.NoError(t, shard1_2.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_2.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 3, shard1_1, shard2_1)
	assertShardValue(ctx, t, 2, shard1_2, shard2_2)
}

func TestJoiningExistingCluster(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancel()

	members := []string{"localhost:51001", "localhost:51002"}

	cluster1 := testCluster(t, members, 1, members[0])
	shard1 := raft.AddShard(ctx, cluster1, 1, &IntStateMachine{})

	cluster2 := testCluster(t, members, 2, members[1])
	shard2 := raft.AddShard(ctx, cluster2, 1, &IntStateMachine{})

	wg, wctx := errgroup.WithContext(ctx)
	wg.Go(func() error { return cluster1.Start(wctx) })
	wg.Go(func() error { return cluster2.Start(wctx) })
	assert.NoError(t, wg.Wait())
	defer cluster1.Stop()
	defer cluster2.Stop()

	// join to the existing cluster as a new member
	cluster3 := testCluster(t, nil, 3, "localhost:51003")
	shard3 := raft.AddShard(ctx, cluster3, 1, &IntStateMachine{})

	assert.NoError(t, cluster1.AddMember(ctx, 1, 3, "localhost:51003"))

	assert.NoError(t, cluster3.Join(ctx))
	defer cluster3.Stop()

	assert.NoError(t, shard3.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 1, shard1, shard2, shard3)

	// join through the new member
	cluster4 := testCluster(t, nil, 4, "localhost:51004")
	shard4 := raft.AddShard(ctx, cluster4, 1, &IntStateMachine{})

	assert.NoError(t, cluster3.AddMember(ctx, 1, 4, "localhost:51004"))
	assert.NoError(t, cluster4.Join(ctx))
	defer cluster4.Stop()

	assert.NoError(t, shard4.Propose(ctx, IntEvent(1)))

	assertShardValue(ctx, t, 2, shard1, shard2, shard3, shard4)
}

func testCluster(t *testing.T, members []string, id uint64, address string) *raft.Cluster {
	return raft.New(&raft.RaftConfig{
		ReplicaID:          id,
		RaftAddress:        address,
		DataDir:            t.TempDir(),
		InitialMembers:     members,
		HeartbeatRTT:       1,
		ElectionRTT:        10,
		SnapshotEntries:    10,
		CompactionOverhead: 10,
	})
}

func assertShardValue(ctx context.Context, t *testing.T, expected int64, shards ...*raft.ShardHandle[IntEvent, int64, int64]) {
	t.Helper()

	for _, shard := range shards {
		res, err := shard.Query(ctx, 0)
		assert.NoError(t, err)
		assert.Equal(t, res, expected)
	}
}
