package ftl

import (
	"context"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

// Topic declares a topic
//
// Topics publish events, and subscriptions can listen to them.
func Topic[E any](name string) TopicHandle[E] {
	return TopicHandle[E]{Ref: &schema.Ref{
		Name:   name,
		Module: reflection.Module(),
	}}
}

type TopicHandle[E any] struct {
	Ref *schema.Ref
}

// Publish publishes an event to a topic
func (t TopicHandle[E]) Publish(ctx context.Context, event E) error {
	return internal.FromContext(ctx).PublishEvent(ctx, t.Ref, event)
}

type SubscriptionHandle[E any] struct {
	Topic *schema.Ref
	Name  string
}

// Subscription declares a subscription to a topic
//
// Sinks can consume events from the subscription by including a "ftl:subscibe <subscription_name>" directive
func Subscription[E any](topic TopicHandle[E], name string) SubscriptionHandle[E] {
	return SubscriptionHandle[E]{Name: name, Topic: topic.Ref}
}
