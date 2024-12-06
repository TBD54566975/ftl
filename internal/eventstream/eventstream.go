package eventstream

import (
	"fmt"
	"sync"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
)

// EventStream is a stream of events that can be published and subscribed to, that update a materialized view
type EventStream[E any, T View[E]] interface {
	Publish(Event[E, T]) error

	View() T

	Subscribe() <-chan Event[E, T]
}

// StreamView is a view of an event stream that can be subscribed to, without modifying the stream.
type StreamView[E any, T View[E]] interface {
	View() T

	Subscribe() <-chan Event[E, T]
}

// View is a read-only view of the materialised current state of the event stream.
type View[E any] interface {
	// Entry returns the entry and the provided address, if it is present in the view
	Entry(string) optional.Option[E]

	// Entries returns all entries in the view
	// these entries are not guaranteed to be in any particular order
	Entries() []E
}

type Event[E any, T View[E]] interface {
	// Address returns the address of the event, which is used to determine where the event should be applied in the view
	Address() string

	// Handle applies the event to the view, by updating the existing entry if it is present
	Handle(optional.Option[E], T) error
}

func NewInMemory[E any, T View[E]](initial T) EventStream[E, T] {
	return &inMemoryEventStream[E, T]{
		view:  initial,
		topic: pubsub.New[Event[E, T]](),
	}

}

type inMemoryEventStream[E any, T View[E]] struct {
	view  T
	lock  sync.Mutex
	topic *pubsub.Topic[Event[E, T]]
}

func (i *inMemoryEventStream[E, T]) Publish(e Event[E, T]) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	address := e.Address()
	existing := i.view.Entry(address)
	err := e.Handle(existing, i.view)
	if err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}
	i.topic.Publish(e)
	return nil
}

func (i *inMemoryEventStream[E, T]) View() T {
	return i.view
}

func (i *inMemoryEventStream[E, T]) Subscribe() <-chan Event[E, T] {
	ret := i.topic.Subscribe(nil)
	return ret
}
