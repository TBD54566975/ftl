package eventstream

import (
	"fmt"
	"sync"

	"github.com/alecthomas/types/pubsub"
)

// EventStream is a stream of events that can be published and subscribed to, that update a materialized view
type EventStream[T View] interface {
	Publish(Event[T]) error

	View() T

	Subscribe() <-chan Event[T]
}

// StreamView is a view of an event stream that can be subscribed to, without modifying the stream.
type StreamView[T View] interface {
	View() T

	Subscribe() <-chan Event[T]
}

// View is a read-only view of the materialised current state of the event stream.
type View interface {
}

type Event[T View] interface {

	// Handle applies the event to the view
	Handle(T) (T, error)
}

func NewInMemory[T View](initial T) EventStream[T] {
	return &inMemoryEventStream[T]{
		view:  initial,
		topic: pubsub.New[Event[T]](),
	}

}

type inMemoryEventStream[T View] struct {
	view  T
	lock  sync.Mutex
	topic *pubsub.Topic[Event[T]]
}

func (i *inMemoryEventStream[T]) Publish(e Event[T]) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	newView, err := e.Handle(i.view)
	if err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}
	i.view = newView
	i.topic.Publish(e)
	return nil
}

func (i *inMemoryEventStream[T]) View() T {
	return i.view
}

func (i *inMemoryEventStream[T]) Subscribe() <-chan Event[T] {
	ret := i.topic.Subscribe(nil)
	return ret
}
