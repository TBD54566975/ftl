package raft_test

import (
	"context"
	"encoding/binary"
	"io"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/raft"
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

	cluster1 := testCluster(t, members, 1)
	shard1_1 := raft.AddShard(ctx, cluster1, 1, &IntStateMachine{})
	shard1_2 := raft.AddShard(ctx, cluster1, 2, &IntStateMachine{})

	cluster2 := testCluster(t, members, 2)
	shard2_1 := raft.AddShard(ctx, cluster2, 1, &IntStateMachine{})
	shard2_2 := raft.AddShard(ctx, cluster2, 2, &IntStateMachine{})
	ready := make(chan struct{})
	go cluster1.Start(ctx, ready) //nolint:errcheck
	go cluster2.Start(ctx, ready) //nolint:errcheck

	<-ready
	<-ready

	assert.NoError(t, shard1_1.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_1.Propose(ctx, IntEvent(2)))
	assert.NoError(t, shard1_2.Propose(ctx, IntEvent(1)))
	assert.NoError(t, shard2_2.Propose(ctx, IntEvent(1)))

	res, err := shard1_1.Query(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, res, int64(3))

	res, err = shard2_1.Query(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, res, int64(3))

	res, err = shard1_2.Query(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, res, int64(2))

	res, err = shard2_2.Query(ctx, 0)
	assert.NoError(t, err)
	assert.Equal(t, res, int64(2))
}

func testCluster(t *testing.T, members []string, id uint64) *raft.Cluster {
	return raft.New(&raft.RaftConfig{
		ReplicaID:          id,
		RaftAddress:        members[id-1],
		DataDir:            t.TempDir(),
		InitialMembers:     members,
		HeartbeatRTT:       1,
		ElectionRTT:        10,
		SnapshotEntries:    10,
		CompactionOverhead: 10,
	})
}
