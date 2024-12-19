package raft

import (
	"context"
	"encoding"
	"fmt"
	"io"

	"github.com/block/ftl/internal/eventstream"
)

type UnitQuery struct{}

type RaftStreamEvent[View encoding.BinaryMarshaler, VPtr Unmarshallable[View]] interface {
	encoding.BinaryMarshaler
	eventstream.Event[View]
}

type RaftEventView[V encoding.BinaryMarshaler, VPrt Unmarshallable[V], E RaftStreamEvent[V, VPrt]] struct {
	shard *ShardHandle[E, UnitQuery, V]
}

func (s *RaftEventView[V, VPrt, E]) Publish(ctx context.Context, event E) error {
	return s.shard.Propose(ctx, event)
}

func (s *RaftEventView[V, VPrt, E]) View(ctx context.Context) (V, error) {
	var zero V

	view, err := s.shard.Query(ctx, UnitQuery{})
	if err != nil {
		return zero, err
	}

	return view, nil
}

type eventStreamStateMachine[
	V encoding.BinaryMarshaler,
	VPrt Unmarshallable[V],
	E RaftStreamEvent[V, VPrt],
	EPtr Unmarshallable[E],
] struct {
	view V
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Close() error {
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Lookup(key UnitQuery) (V, error) {
	return s.view, nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Update(msg E) error {
	v, err := msg.Handle(s.view)
	if err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}
	s.view = v
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Save(writer io.Writer) error {
	bytes, err := s.view.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal view: %w", err)
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write view: %w", err)
	}
	return nil
}

func (s *eventStreamStateMachine[V, VPrt, E, EPtr]) Recover(reader io.Reader) error {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read view: %w", err)
	}
	if err := (VPrt)(&s.view).UnmarshalBinary(bytes); err != nil {
		return fmt.Errorf("failed to unmarshal view: %w", err)
	}
	return nil
}

func NewRaftEventStream[
	V encoding.BinaryMarshaler,
	VPtr Unmarshallable[V],
	E RaftStreamEvent[V, VPtr],
	EPtr Unmarshallable[E],
](ctx context.Context, cluster *Cluster, shardID uint64) eventstream.EventView[V, E] {
	sm := &eventStreamStateMachine[V, VPtr, E, EPtr]{}
	shard := AddShard[UnitQuery, V, E, EPtr](ctx, cluster, shardID, sm)
	return &RaftEventView[V, VPtr, E]{shard: shard}
}
