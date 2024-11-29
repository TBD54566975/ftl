package ftl

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/schema"
)

// TopicPartitionMap maps an event to a partition key
type TopicPartitionMap[E any] interface {
	PartitionKey(event E) string
}

// SinglePartitionMap can be used for topics with a single partition
type SinglePartitionMap[E any] struct{}

var _ TopicPartitionMap[struct{}] = SinglePartitionMap[struct{}]{}

func (SinglePartitionMap[E]) PartitionKey(_ E) string { return "" }

// TopicHandle accesses a topic
//
// Topics publish events, and subscriptions can listen to them.
type TopicHandle[E any, M TopicPartitionMap[E]] struct {
	Ref          *schema.Ref
	PartitionMap M
}

// Publish publishes an event to a topic
func (t TopicHandle[E, M]) Publish(ctx context.Context, event E) error {
	var mapper M
	return internal.FromContext(ctx).PublishEvent(ctx, t.Ref, event, mapper.PartitionKey(event)) //nolint:wrapcheck
}
