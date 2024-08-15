package pubsub

import (
	"context"
	"strconv"
	"testing"

	"github.com/alecthomas/atomic"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestSubscriberReturningErrors(t *testing.T) {
	// Test that we can publish multiple events, which will take time to consume, and that we track each error
	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription1_1, ErrorsAfterASecond),
	)
	count := 5
	for i := 0; i < count; i++ {
		assert.NoError(t, PublishToTopicOne(ctx, Event{Value: strconv.Itoa(i)}))
	}
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, count, len(ftltest.ErrorsForSubscription(ctx, subscription1_1)))
	assert.Equal(t, count, len(ftltest.EventsForTopic(ctx, Topic1)))
}

// establishes a pubsub network that forwards from topic 1 to topic 2 on a single subscription
// and does NOT register any subscribers against topic 2's subscription
func TestForwardedEvent(t *testing.T) {
	// Test that we can publish multiple events, which will take time to consume, and that we track each error
	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription1_1, PropagateToTopic2),
	)
	assert.NoError(t, PublishToTopicOne(ctx, Event{Value: "propagation-test"}))
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic1)))
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic2)))
}

// establishes a pubsub network that forwards from topic 1 to topic 2 on two subscriptions
// and does NOT register any subscribers against topic 2's subscriptions
func TestPropagatedEvent(t *testing.T) {
	// Test that we can publish multiple events, which will take time to consume, and that we track each error
	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription1_1, PropagateToTopic2),
		ftltest.WithSubscriber(subscription1_2, PropagateToTopic2),
	)
	assert.NoError(t, PublishToTopicOne(ctx, Event{Value: "propagation-test"}))
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic1)))
	assert.Equal(t, 2, len(ftltest.EventsForTopic(ctx, Topic2)))
}

// establishes a pubsub network that forwards from topic 1 to topic 2 on two subscriptions
// and consumes from topic 2 via two subscriptions
func TestPropagationNetwork(t *testing.T) {
	// Test that we can publish multiple events, which will take time to consume, and that we track each error
	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription1_1, PropagateToTopic2),
		ftltest.WithSubscriber(subscription1_2, PropagateToTopic2),
		ftltest.WithSubscriber(subscription2_1, ConsumeEvent),
		ftltest.WithSubscriber(subscription2_2, ConsumeEvent),
	)
	assert.NoError(t, PublishToTopicOne(ctx, Event{Value: "propagation-test"}))
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic1)))
	assert.Equal(t, 2, len(ftltest.EventsForTopic(ctx, Topic2)))
}

func TestMultipleMultipleFakeSubscribers(t *testing.T) {
	// Test that multiple adhoc subscribers can be added for a subscription
	count := 5
	var counter atomic.Value[int]

	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription1_1, func(ctx context.Context, event Event) error {
			ftl.LoggerFromContext(ctx).Infof("Fake Subscriber A")
			current := counter.Load()
			counter.Store(current + 1)
			return nil
		}),
		ftltest.WithSubscriber(subscription1_1, func(ctx context.Context, event Event) error {
			ftl.LoggerFromContext(ctx).Infof("Fake Subscriber B")
			current := counter.Load()
			counter.Store(current + 1)
			return nil
		}),
	)
	for i := 0; i < count; i++ {
		assert.NoError(t, PublishToTopicOne(ctx, Event{Value: strconv.Itoa(i)}))
	}
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, 0, len(ftltest.ErrorsForSubscription(ctx, subscription1_1)))
	assert.Equal(t, count, len(ftltest.EventsForTopic(ctx, Topic1)))
	assert.Equal(t, count, counter.Load())
}
