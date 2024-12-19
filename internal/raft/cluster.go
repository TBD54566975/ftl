package raft

import (
	"context"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
	"github.com/lni/dragonboat/v4"
	"github.com/lni/dragonboat/v4/client"
	"github.com/lni/dragonboat/v4/config"
	"github.com/lni/dragonboat/v4/statemachine"
)

type RaftConfig struct {
	InitialMembers    []string      `help:"Initial members" required:""`
	ReplicaID         uint64        `help:"Node ID" required:""`
	DataDir           string        `help:"Data directory" required:""`
	RaftAddress       string        `help:"Address to advertise to other nodes" required:""`
	ListenAddress     string        `help:"Address to listen for incoming traffic. If empty, RaftAddress will be used."`
	ShardReadyTimeout time.Duration `help:"Timeout for shard to be ready" default:"5s"`
	// Raft configuration
	RTT                time.Duration `help:"Estimated average round trip time between nodes" default:"200ms"`
	ElectionRTT        uint64        `help:"Election RTT as a multiple of RTT" default:"10"`
	HeartbeatRTT       uint64        `help:"Heartbeat RTT as a multiple of RTT" default:"1"`
	SnapshotEntries    uint64        `help:"Snapshot entries" default:"10"`
	CompactionOverhead uint64        `help:"Compaction overhead" default:"100"`
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
// This can be only called before the cluster is started.
func AddShard[Q any, R any, E Event, EPtr Unmarshallable[E]](
	ctx context.Context,
	to *Cluster,
	shardID uint64,
	sm StateMachine[Q, R, E, EPtr],
) *ShardHandle[E, Q, R] {
	if to.nh != nil {
		panic("cluster already started")
	}
	to.shards[shardID] = newStateMachineShim[Q, R, E, EPtr](sm)
	return &ShardHandle[E, Q, R]{
		shardID: shardID,
		cluster: to,
	}
}

// Start the cluster. Blocks until the cluster instance is ready.
func (c *Cluster) Start(ctx context.Context) error {
	return c.start(ctx, false)
}

// Join the cluster as a new member. Blocks until the cluster instance is ready.
func (c *Cluster) Join(ctx context.Context) error {
	return c.start(ctx, true)
}

func (c *Cluster) start(ctx context.Context, join bool) error {
	// Create node host config
	nhc := config.NodeHostConfig{
		WALDir:         c.config.DataDir,
		NodeHostDir:    c.config.DataDir,
		RTTMillisecond: uint64(c.config.RTT.Milliseconds()),
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
		if !join {
			for idx, peer := range c.config.InitialMembers {
				peers[uint64(idx+1)] = peer
			}
		}

		// Start the raft node for this shard
		if err := nh.StartReplica(peers, join, sm, cfg); err != nil {
			return fmt.Errorf("failed to start replica %d for shard %d: %w", c.config.ReplicaID, shardID, err)
		}
	}

	// Wait for all shards to be ready
	// TODO: WaitReady in the config should do this, but for some reason it doesn't work.
	for shardID := range c.shards {
		if err := c.waitReady(ctx, shardID); err != nil {
			return fmt.Errorf("failed to wait for shard %d to be ready on replica %d: %w", shardID, c.config.ReplicaID, err)
		}
	}

	return nil
}

func (c *Cluster) Stop() {
	if c.nh == nil {
		return
	}
	c.nh.Close()
	c.nh = nil
}

// AddMember to the cluster. This needs to be called on an existing running cluster member,
// before the new member is started.
func (c *Cluster) AddMember(ctx context.Context, shardID uint64, replicaID uint64, address string) error {
	if err := c.nh.SyncRequestAddReplica(ctx, shardID, replicaID, address, 0); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (c *Cluster) waitReady(ctx context.Context, shardID uint64) error {
	retry := backoff.Backoff{
		Min:    5 * time.Millisecond,
		Max:    c.config.ShardReadyTimeout,
		Factor: 2,
		Jitter: true,
	}
	for {
		rs, err := c.nh.ReadIndex(shardID, c.config.ShardReadyTimeout)
		if err != nil || rs == nil {
			return fmt.Errorf("failed to read index: %w", err)
		}
		res := <-rs.ResultC()
		rs.Release()
		if !res.Completed() {
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled")
			case <-time.After(retry.Duration()):
			}
			continue
		}
		break
	}
	return nil
}
