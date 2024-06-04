package ftl

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/internal"
)

// Topic declares a topic
//
// Topics publish events, and subscriptions can listen to them.
func Topic[E any](name string) TopicHandle[E] {
	return TopicHandle[E]{name: name}
}

type TopicHandle[E any] struct {
	name string
}

// Publish publishes an event to a topic
func (t TopicHandle[E]) Publish(ctx context.Context, event E) error {
	return internal.FromContext(ctx).PublishEvent(ctx, t.name, event)
}

// Subscription declares a subscription to a topic
//
// Sinks can consume events from the subscription by including a "ftl:subscibe <subscription_name>" directive
func Subscription[E any](topic TopicHandle[E], name string) SubscriptionHandle[E] {
	return SubscriptionHandle[E]{name: name}
}

type SubscriptionHandle[E any] struct {
	name string
}
