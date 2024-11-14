package ftl

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/schema"
)

// TopicHandle accesses a topic
//
// Topics publish events, and subscriptions can listen to them.
type TopicHandle[E any] struct {
	Ref *schema.Ref
}

// Publish publishes an event to a topic
func (t TopicHandle[E]) Publish(ctx context.Context, event E) error {
	return internal.FromContext(ctx).PublishEvent(ctx, t.Ref, event)
}

// SubscriptionHandle declares a subscription to a topic for the provided Sink
// T: the topic handle type
// S: the generated sink client type
// E: the event type
type SubscriptionHandle[T TopicHandle[E], S, E any] struct {
}
