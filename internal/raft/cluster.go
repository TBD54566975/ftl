package raft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/statemachine"
	"github.com/lni/goutils/syncutil"
)

type RaftConfig struct {
	InitialMembers []string `help:"Initial members" required:""`
	ReplicaID      uint64   `help:"Node ID" required:""`
	DataDir        string   `help:"Data directory" required:""`
	RaftAddress    string   `help:"Address to advertise to other nodes" required:""`
	ListenAddress  string   `help:"Address to listen for incoming traffic. If empty, RaftAddress will be used."`
	// Raft configuration
	ElectionRTT        uint64 `help:"Election RTT" default:"10"`
	HeartbeatRTT       uint64 `help:"Heartbeat RTT" default:"1"`
	SnapshotEntries    uint64 `help:"Snapshot entries" default:"10"`
	CompactionOverhead uint64 `help:"Compaction overhead" default:"100"`
}

// Cluster of dragonboat nodes.
type Cluster struct {
	config *RaftConfig
	nh     *dragonboat.NodeHost
	shards map[uint64]statemachine.CreateStateMachineFunc
}

// ShardHandle is a handle to a shard in the cluster.
// It is the interface to update and query the state of a shard.
//
// E is the event type.
// Q is the query type.
// R is the query response type.
type ShardHandle[E Event, Q any, R any] struct {
	shardID uint64
	cluster *Cluster
	session *client.Session
}

// Propose an event to the shard.
func (s *ShardHandle[E, Q, R]) Propose(ctx context.Context, msg E) error {
	if s.cluster.nh == nil {
		panic("cluster not started")
	}

	msgBytes, err := msg.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	if s.session == nil {
		// use a no-op session for now. This means that a retry on timeout could result into duplicate events.
		s.session = s.cluster.nh.GetNoOPSession(s.shardID)
	}
	_, err = s.cluster.nh.SyncPropose(ctx, s.session, msgBytes)
	if err != nil {
		return fmt.Errorf("failed to propose event: %w", err)
	}
	return nil
}

// Query the state of the shard.
func (s *ShardHandle[E, Q, R]) Query(ctx context.Context, query Q) (R, error) {
	if s.cluster.nh == nil {
		panic("cluster not started")
	}

	var zero R

	res, err := s.cluster.nh.SyncRead(ctx, s.shardID, query)
	if err != nil {
		return zero, fmt.Errorf("failed to query shard: %w", err)
	}

	response, ok := res.(R)
	if !ok {
		panic(fmt.Errorf("invalid response type: %T", res))
	}

	return response, nil
}

// New creates a new cluster.
func New(cfg *RaftConfig) *Cluster {
	return &Cluster{
		config: cfg,
		shards: make(map[uint64]statemachine.CreateStateMachineFunc),
	}
}

// AddShard adds a shard to the cluster.
func AddShard[Q any, R any, E Event, EPtr Unmasrshallable[E]](
	ctx context.Context,
	to *Cluster,
	shardID uint64,
	sm StateMachine[Q, R, E, EPtr],
) *ShardHandle[E, Q, R] {
	to.shards[shardID] = newStateMachineShim[Q, R, E, EPtr](sm)
	return &ShardHandle[E, Q, R]{
		shardID: shardID,
		cluster: to,
	}
}

// Start the cluster.
func (c *Cluster) Start(ctx context.Context, ready chan struct{}) error {
	// Create node host config
	nhc := config.NodeHostConfig{
		WALDir:         c.config.DataDir,
		NodeHostDir:    c.config.DataDir,
		RTTMillisecond: 200,
		RaftAddress:    c.config.RaftAddress,
		ListenAddress:  c.config.ListenAddress,
	}

	// Create node host
	nh, err := dragonboat.NewNodeHost(nhc)
	if err != nil {
		return fmt.Errorf("failed to create node host: %w", err)
	}
	c.nh = nh

	// Start replicas for each shard
	for shardID, sm := range c.shards {
		cfg := config.Config{
			ReplicaID:          c.config.ReplicaID,
			ShardID:            shardID,
			CheckQuorum:        true,
			ElectionRTT:        c.config.ElectionRTT,
			HeartbeatRTT:       c.config.HeartbeatRTT,
			SnapshotEntries:    c.config.SnapshotEntries,
			CompactionOverhead: c.config.CompactionOverhead,
			WaitReady:          true,
		}

		peers := make(map[uint64]string)
		for idx, peer := range c.config.InitialMembers {
			peers[uint64(idx+1)] = peer
		}

		// Start the raft node for this shard
		if err := nh.StartReplica(peers, false, sm, cfg); err != nil {
			return fmt.Errorf("failed to start replica for shard %d: %w", shardID, err)
		}
	}

	raftstopper := syncutil.NewStopper()
	raftstopper.RunWorker(func() {
		<-ctx.Done()
		c.nh.Close()
	})

	// Wait for all shards to be ready
	// TODO: WaitReady in the config should do this, but for some reason it doesn't work.
	for shardID := range c.shards {
		if err := c.waitReady(ctx, shardID); err != nil {
			return fmt.Errorf("failed to wait for shard %d to be ready: %w", shardID, err)
		}
	}

	if ready != nil {
		ready <- struct{}{}
	}
	raftstopper.Wait()

	for shardID := range c.shards {
		if err := c.nh.StopReplica(shardID, c.config.ReplicaID); err != nil {
			return fmt.Errorf("failed to stop replica for shard %d: %w", shardID, err)
		}
	}

	return nil
}

func (c *Cluster) waitReady(ctx context.Context, shardID uint64) error {
	retry := backoff.Backoff{}
	for {
		_, err := c.nh.SyncGetShardMembership(ctx, shardID)
		if err == nil {
			return nil
		}
		if !errors.Is(err, dragonboat.ErrShardNotReady) {
			return fmt.Errorf("failed to get shard membership: %w", err)
		}
		time.Sleep(retry.Duration())
	}
}
