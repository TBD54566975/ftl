package eventstream

import (
	"fmt"
	"sync"

	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl/internal/reflect"
)

// EventStream is a stream of events that can be published and subscribed to, that update a materialized view
type EventStream[View any] interface {
	Publish(event Event[View]) error

	View() View

	Subscribe() <-chan Event[View]
}

// StreamView is a view of an event stream that can be subscribed to, without modifying the stream.
type StreamView[View any] interface {
	View() View

	// Subscribe to the event stream. The channel will only receive events that are published after the subscription.
	Subscribe() <-chan Event[View]
}

type Event[View any] interface {

	// Handle applies the event to the view
	Handle(view View) (View, error)
}

func NewInMemory[View any](initial View) EventStream[View] {
	return &inMemoryEventStream[View]{
		view:  initial,
		topic: pubsub.New[Event[View]](),
	}
}

type inMemoryEventStream[View any] struct {
	view  View
	lock  sync.Mutex
	topic *pubsub.Topic[Event[View]]
}

func (i *inMemoryEventStream[T]) Publish(e Event[T]) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	newView, err := e.Handle(reflect.DeepCopy(i.view))
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
