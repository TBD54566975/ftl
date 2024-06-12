package ftltest

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
)

type fakePubSub struct {
	// all pubsub events are processed through globalTopic
	globalTopic *pubsub.Topic[pubSubEvent]
	// publishWaitGroup can be used to wait for all events to be published
	publishWaitGroup sync.WaitGroup

	// pubSubLock required to access [topics, subscriptions, subscribers]
	pubSubLock    sync.Mutex
	topics        map[schema.RefKey][]any
	subscriptions map[string]*subscription
	subscribers   map[string][]subscriber
}

func newFakePubSub(ctx context.Context) *fakePubSub {
	f := &fakePubSub{
		globalTopic:   pubsub.New[pubSubEvent](),
		topics:        map[schema.RefKey][]any{},
		subscriptions: map[string]*subscription{},
		subscribers:   map[string][]subscriber{},
	}
	f.watchPubSub(ctx)
	return f
}

func (f *fakePubSub) publishEvent(topic *schema.Ref, event any) error {
	f.publishWaitGroup.Add(1)
	return f.globalTopic.PublishSync(publishEvent{topic: topic, content: event})
}

// addSubscriber adds a subscriber to the fake FTL instance. Each subscriber included in the test must be manually added
func addSubscriber[E any](f *fakePubSub, sub ftl.SubscriptionHandle[E], sink ftl.Sink[E]) {
	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	if _, ok := f.subscriptions[sub.Name]; !ok {
		f.subscriptions[sub.Name] = &subscription{
			name:   sub.Name,
			topic:  sub.Topic,
			errors: map[int]error{},
		}
	}

	f.subscribers[sub.Name] = append(f.subscribers[sub.Name], func(ctx context.Context, event any) error {
		if event, ok := event.(E); ok {
			return sink(ctx, event)
		}
		return fmt.Errorf("unexpected event type %T for subscription %s", event, sub.Name)
	})
}

// eventsForTopic returns all events published to a topic
func eventsForTopic[E any](ctx context.Context, f *fakePubSub, topic ftl.TopicHandle[E]) []E {
	// Make sure all published events make it into our pubsub state
	f.publishWaitGroup.Wait()

	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	logger := log.FromContext(ctx).Scope("pubsub")
	var events = []E{}
	raw, ok := f.topics[topic.Ref.ToRefKey()]
	if !ok {
		return events
	}
	for _, e := range raw {
		if event, ok := e.(E); ok {
			events = append(events, event)
		} else {
			logger.Warnf("unexpected event type %T for topic %s", e, topic.Ref)
		}
	}
	return events
}

// resultsForSubscription returns all consumed events and whether an error was returned
func resultsForSubscription[E any](ctx context.Context, f *fakePubSub, handle ftl.SubscriptionHandle[E]) []SubscriptionResult[E] {
	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	logger := log.FromContext(ctx).Scope("pubsub")
	results := []SubscriptionResult[E]{}

	subscription, ok := f.subscriptions[handle.Name]
	if !ok {
		return results
	}
	topic, ok := f.topics[handle.Topic.ToRefKey()]
	if !ok {
		return results
	}

	count := subscription.cursor.Default(-1)
	if !subscription.isExecuting {
		count++
	}
	for i := range count {
		e := topic[i]
		if event, ok := e.(E); ok {
			result := SubscriptionResult[E]{
				Event: event,
			}
			if err, ok := subscription.errors[i]; ok {
				result.Error = ftl.Some(err)
			}
			results = append(results, result)
		} else {
			logger.Warnf("unexpected event type %T for subscription %s", e, handle.Name)
		}

	}
	return results
}

func (f *fakePubSub) watchPubSub(ctx context.Context) {
	events := make(chan pubSubEvent, 128)
	f.globalTopic.Subscribe(events)
	go func() {
		defer f.globalTopic.Unsubscribe(events)
		for {
			select {
			case e := <-events:
				f.handlePubSubEvent(ctx, e)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (f *fakePubSub) handlePubSubEvent(ctx context.Context, e pubSubEvent) {
	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	logger := log.FromContext(ctx).Scope("pubsub")

	switch event := e.(type) {
	case publishEvent:
		logger.Debugf("publishing to %s: %v", event.topic.Name, event.content)
		if _, ok := f.topics[event.topic.ToRefKey()]; !ok {
			f.topics[event.topic.ToRefKey()] = []any{event.content}
		} else {
			f.topics[event.topic.ToRefKey()] = append(f.topics[event.topic.ToRefKey()], event.content)
		}
		f.publishWaitGroup.Done()
	case subscriptionDidConsumeEvent:
		sub, ok := f.subscriptions[event.subscription]
		if !ok {
			panic(fmt.Sprintf("subscription %q not found", event.subscription))
		}
		if event.err != nil {
			sub.errors[sub.cursor.MustGet()] = event.err
		}
		sub.isExecuting = false
	}

	for _, sub := range f.subscriptions {
		if sub.isExecuting {
			// already executing
			continue
		}
		topicEvents, ok := f.topics[sub.topic.ToRefKey()]
		if !ok {
			// no events publshed yet
			continue
		}
		var cursor = sub.cursor.Default(-1)
		if len(topicEvents) <= cursor+1 {
			// no new events
			continue
		}
		subscribers, ok := f.subscribers[sub.name]
		if !ok || len(subscribers) == 0 {
			// no subscribers
			continue
		}
		chosenSubscriber := subscribers[rand.Intn(len(subscribers))] //nolint:gosec

		sub.cursor = optional.Some(cursor + 1)
		sub.isExecuting = true

		go func(sub string, chosenSubscriber subscriber, event any) {
			err := chosenSubscriber(ctx, event)
			f.globalTopic.Publish(subscriptionDidConsumeEvent{subscription: sub, err: err})
		}(sub.name, chosenSubscriber, topicEvents[sub.cursor.MustGet()])
	}
}

// waitForSubscriptionsToComplete waits for all subscriptions to consume all events.
// subscriptions with no subscribers are ignored.
// logs what which subscriptions are blocking every 2s.
func (f *fakePubSub) waitForSubscriptionsToComplete(ctx context.Context) {
	startTime := time.Now()
	nextLogTime := startTime.Add(2 * time.Second)

	// Make sure all published events make it into our pubsub state
	f.publishWaitGroup.Wait()

	for {
		shouldPrint := time.Now().After(nextLogTime)
		if f.checkSubscriptionsAreComplete(ctx, shouldPrint, startTime) {
			return
		}
		if shouldPrint {
			nextLogTime = time.Now().Add(2 * time.Second)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (f *fakePubSub) checkSubscriptionsAreComplete(ctx context.Context, shouldPrint bool, startTime time.Time) bool {
	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	type remainingState struct {
		name          string
		isExecuting   bool
		pendingEvents int
	}
	remaining := []remainingState{}
	for _, sub := range f.subscriptions {
		topicEvents, ok := f.topics[sub.topic.ToRefKey()]
		if !ok {
			// no events publshed yet
			continue
		}
		var cursor = sub.cursor.Default(-1)
		if !sub.isExecuting && len(topicEvents) <= cursor+1 {
			// all events have been consumed
			continue
		}
		subscribers, ok := f.subscribers[sub.name]
		if !ok || len(subscribers) == 0 {
			// no subscribers
			continue
		}
		remaining = append(remaining, remainingState{
			name:          sub.name,
			isExecuting:   sub.isExecuting,
			pendingEvents: len(topicEvents) - cursor - 1,
		})
	}
	if len(remaining) == 0 {
		// not waiting on any more subscriptions
		return true
	}
	if shouldPrint {
		// print out what we are waiting on
		logger := log.FromContext(ctx).Scope("pubsub")
		logger.Debugf("waiting on subscriptions to complete after %ds:\n%s", int(time.Until(startTime).Seconds()*-1), strings.Join(slices.Map(remaining, func(r remainingState) string {
			var suffix string
			if r.isExecuting {
				suffix = ", 1 executing"
			}
			return fmt.Sprintf("  %s: %d events pending%s", r.name, r.pendingEvents, suffix)
		}), "\n"))
	}
	return false
}
