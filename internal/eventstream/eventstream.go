package eventstream

import (
	"context"
	"fmt"
	"sync"

	"github.com/alecthomas/types/pubsub"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/reflect"
)

// EventStream is a stream of events that can be published and subscribed to, that update a materialized view
type EventStream[View any, E Event[View]] interface {
	Publish(ctx context.Context, event E) error

	View() View

	Updates() *pubsub.Topic[E]
}

// StreamView is a view of an event stream that can be subscribed to, without modifying the stream.
type StreamView[View any] interface {
	View() View

	// Subscribe to the event stream. The channel will only receive events that are published after the subscription.
	Subscribe(ctx context.Context) <-chan Event[View]
}

type Event[View any] interface {

	// Handle applies the event to the view
	Handle(view View) (View, error)
}

func NewInMemory[View any, E Event[View]](initial View) EventStream[View, E] {
	return &inMemoryEventStream[View, E]{
		view:  initial,
		topic: pubsub.New[E](),
	}
}

type inMemoryEventStream[View any, E Event[View]] struct {
	view  View
	lock  sync.Mutex
	topic *pubsub.Topic[E]
}

func (i *inMemoryEventStream[T, E]) Publish(ctx context.Context, e E) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	logger := log.FromContext(ctx)

	if _, ok := any(e).(VerboseMessage); ok {
		logger.Tracef("Publishing event %T%v", e, e)
	} else {
		logger.Debugf("Publishing event %T%v", e, e)
	}
	newView, err := e.Handle(reflect.DeepCopy(i.view))
	if err != nil {
		return fmt.Errorf("failed to handle event: %w", err)
	}
	i.view = newView
	i.topic.Publish(e)
	return nil
}

func (i *inMemoryEventStream[T, E]) View() T {
	return i.view
}

func (i *inMemoryEventStream[T, E]) Updates() *pubsub.Topic[E] {
	return i.topic
}

type VerboseMessage interface {
	VerboseMessage()
}
