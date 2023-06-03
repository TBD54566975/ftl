package eventsource

import (
	"github.com/alecthomas/atomic"

	"github.com/TBD54566975/ftl/internal/pubsub"
)

// EventSource is a pubsub.Topic that also stores the last published value in an atomic.Value.
//
// Updating the value will result in a publish event.
type EventSource[T any] struct {
	*pubsub.Topic[T]
	value *atomic.Value[T]
}

var _ atomic.Interface[int] = (*EventSource[int])(nil)

func New[T any]() *EventSource[T] {
	var t T
	e := &EventSource[T]{Topic: pubsub.New[T](), value: atomic.New(t)}
	changes := make(chan T, 64)
	e.Subscribe(changes)
	go func() {
		for value := range changes {
			e.value.Store(value)
		}
	}()
	return e
}

func (e *EventSource[T]) Store(value T) {
	e.value.Store(value)
	e.Publish(value)
}

func (e *EventSource[T]) Load() T {
	return e.value.Load()
}

func (e *EventSource[T]) Swap(value T) T {
	rv := e.value.Swap(value)
	e.Publish(value)
	return rv
}

func (e *EventSource[T]) CompareAndSwap(old, new T) bool { //nolint:predeclared
	if e.value.CompareAndSwap(old, new) {
		e.Publish(new)
		return true
	}
	return false
}
