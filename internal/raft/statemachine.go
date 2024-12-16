package raft

import (
	"encoding"
	"fmt"
	"io"

	"github.com/lni/dragonboat/v4/statemachine"
)

// Event to update the state machine. These are stored in the Raft log.
type Event interface {
	encoding.BinaryMarshaler
}

// Unmasrshallable is a type that can be unmarshalled from a binary representation.
type Unmasrshallable[T any] interface {
	*T
	encoding.BinaryUnmarshaler
}

// StateMachine is a typed interface to dragonboat's statemachine.IStateMachine.
// It is used to implement the state machine for a single shard.
//
// Q is the query type.
// R is the query response type.
// E is the event type.
type StateMachine[Q any, R any, E Event, EPtr Unmasrshallable[E]] interface {
	// Query the state of the state machine.
	Lookup(key Q) (R, error)
	// Update the state of the state machine.
	Update(msg E) error
	// Save the state of the state machine to a snapshot.
	Save(writer io.Writer) error
	// Recover the state of the state machine from a snapshot.
	Recover(reader io.Reader) error
	// Close the state machine.
	Close() error
}

type stateMachineShim[Q any, R any, E Event, EPtr Unmasrshallable[E]] struct {
	sm StateMachine[Q, R, E, EPtr]
}

func newStateMachineShim[Q any, R any, E Event, EPtr Unmasrshallable[E]](
	sm StateMachine[Q, R, E, EPtr],
) statemachine.CreateStateMachineFunc {
	return func(clusterID uint64, nodeID uint64) statemachine.IStateMachine {
		return &stateMachineShim[Q, R, E, EPtr]{sm: sm}
	}
}

func (s *stateMachineShim[Q, R, E, EPtr]) Lookup(key any) (any, error) {
	typed, ok := key.(Q)
	if !ok {
		panic(fmt.Errorf("invalid key type: %T", key))
	}

	res, err := s.sm.Lookup(typed)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup: %w", err)
	}

	return res, nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) Update(entry statemachine.Entry) (statemachine.Result, error) {
	var to E
	toptr := (EPtr)(&to)

	if err := toptr.UnmarshalBinary(entry.Cmd); err != nil {
		return statemachine.Result{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	if err := s.sm.Update(to); err != nil {
		return statemachine.Result{}, fmt.Errorf("failed to update state machine: %w", err)
	}

	return statemachine.Result{}, nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) Close() error {
	if err := s.sm.Close(); err != nil {
		return fmt.Errorf("failed to close state machine: %w", err)
	}
	return nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) RecoverFromSnapshot(
	reader io.Reader,
	_ []statemachine.SnapshotFile, // do not support extra immutable files for now
	_ <-chan struct{}, // do not support snapshot recovery cancellation for now
) error {
	if err := s.sm.Recover(reader); err != nil {
		return fmt.Errorf("failed to recover from snapshot: %w", err)
	}
	return nil
}

func (s *stateMachineShim[Q, R, E, EPtr]) SaveSnapshot(
	writer io.Writer,
	_ statemachine.ISnapshotFileCollection, // do not support extra immutable files for now
	_ <-chan struct{}, // do not support snapshot save cancellation for now
) error {
	if err := s.sm.Save(writer); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}
	return nil
}
