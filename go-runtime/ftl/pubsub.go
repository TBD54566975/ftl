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
