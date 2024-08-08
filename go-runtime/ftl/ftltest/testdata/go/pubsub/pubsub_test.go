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
		ftltest.WithSubscriber(subscription, ErrorsAfterASecond),
	)
	count := 5
	for i := 0; i < count; i++ {
		assert.NoError(t, PublishToTopicOne(ctx, Event{Value: strconv.Itoa(i)}))
	}
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, count, len(ftltest.ErrorsForSubscription(ctx, subscription)))
	assert.Equal(t, count, len(ftltest.EventsForTopic(ctx, Topic)))
}

func TestPropagateEvents(t *testing.T) {
	// Test that we can publish multiple events, which will take time to consume, and that we track each error
	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription, PublishToTopicTwo),
	)
	assert.NoError(t, PublishToTopicOne(ctx, Event{Value: "propagation-test"}))
	ftltest.WaitForSubscriptionsToComplete(ctx)
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic)))
	assert.Equal(t, 1, len(ftltest.EventsForTopic(ctx, Topic2)))
}

func TestMultipleMultipleFakeSubscribers(t *testing.T) {
	// Test that multiple adhoc subscribers can be added for a subscription
	count := 5
	var counter atomic.Value[int]

	ctx := ftltest.Context(
		ftltest.WithSubscriber(subscription, func(ctx context.Context, event Event) error {
			ftl.LoggerFromContext(ctx).Infof("Fake Subscriber A")
			current := counter.Load()
			counter.Store(current + 1)
			return nil
		}),
		ftltest.WithSubscriber(subscription, func(ctx context.Context, event Event) error {
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
	assert.Equal(t, 0, len(ftltest.ErrorsForSubscription(ctx, subscription)))
	assert.Equal(t, count, len(ftltest.EventsForTopic(ctx, Topic)))
	assert.Equal(t, count, counter.Load())
}
