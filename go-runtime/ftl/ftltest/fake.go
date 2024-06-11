package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
)

// pubSubEvent is a sum type for all events that can be published to the pubsub system.
// not to be confused with an event that gets published to a topic
//
//sumtype:decl
type pubSubEvent interface {
	// cronJobEvent is a marker to ensure that all events implement the interface.
	pubSubEvent()
}

// publishEvent holds an event to be published to a topic
type publishEvent struct {
	topic   *schema.Ref
	content any
}

func (publishEvent) pubSubEvent() {}

// subscriptionDidConsumeEvent indicates that a call to a subscriber has completed
type subscriptionDidConsumeEvent struct {
	subscription string
	err          error
}

func (subscriptionDidConsumeEvent) pubSubEvent() {}

type subscription struct {
	name        string
	topic       *schema.Ref
	cursor      optional.Option[int]
	isExecuting bool
	errors      map[int]error
}

type subscriber func(context.Context, any) error

type fakeFTL struct {
	fsm *fakeFSMManager

	mockMaps      map[uintptr]mapImpl
	allowMapCalls bool
	configValues  map[string][]byte
	secretValues  map[string][]byte

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

// mapImpl is a function that takes an object and returns an object of a potentially different
// type but is not constrained by input/output type like ftl.Map.
type mapImpl func(context.Context) (any, error)

func newFakeFTL(ctx context.Context) *fakeFTL {
	fake := &fakeFTL{
		fsm:           newFakeFSMManager(),
		mockMaps:      map[uintptr]mapImpl{},
		allowMapCalls: false,
		configValues:  map[string][]byte{},
		secretValues:  map[string][]byte{},
		globalTopic:   pubsub.New[pubSubEvent](),
		topics:        map[schema.RefKey][]any{},
		subscriptions: map[string]*subscription{},
		subscribers:   map[string][]subscriber{},
	}

	fake.watchPubSub(ctx)

	return fake
}

var _ internal.FTL = &fakeFTL{}

func (f *fakeFTL) setConfig(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.configValues[name] = data
	return nil
}

func (f *fakeFTL) GetConfig(ctx context.Context, name string, dest any) error {
	data, ok := f.configValues[name]
	if !ok {
		return fmt.Errorf("secret value %q not found: %w", name, configuration.ErrNotFound)
	}
	return json.Unmarshal(data, dest)
}

func (f *fakeFTL) setSecret(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.secretValues[name] = data
	return nil
}

func (f *fakeFTL) GetSecret(ctx context.Context, name string, dest any) error {
	data, ok := f.secretValues[name]
	if !ok {
		return fmt.Errorf("config value %q not found: %w", name, configuration.ErrNotFound)
	}
	return json.Unmarshal(data, dest)
}

func (f *fakeFTL) FSMSend(ctx context.Context, fsm string, instance string, event any) error {
	return f.fsm.SendEvent(ctx, fsm, instance, event)
}

// addMapMock saves a new mock of ftl.Map to the internal map in fakeFTL.
//
// mockMap provides the whole mock implemention, so it gets called in place of both `fn`
// and `getter` in ftl.Map.
func (f *fakeFTL) addMapMock(mapper any, mockMap func(context.Context) (any, error)) {
	key := makeMapKey(mapper)
	f.mockMaps[key] = mockMap
}

func (f *fakeFTL) startAllowingMapCalls() {
	f.allowMapCalls = true
}

func (f *fakeFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	key := makeMapKey(mapper)
	mockMap, ok := f.mockMaps[key]
	if ok {
		return actuallyCallMap(ctx, mockMap)
	}
	if f.allowMapCalls {
		return actuallyCallMap(ctx, mapImpl)
	}
	panic("map calls not allowed in tests by default. ftltest.Context should be instantiated with either ftltest.WithMapsAllowed() or a mock for the specific map being called using ftltest.WhenMap(...)")
}

func makeMapKey(mapper any) uintptr {
	v := reflect.ValueOf(mapper)
	if v.Kind() != reflect.Pointer {
		panic("fakeFTL received object that was not a pointer, expected *MapHandle")
	}
	underlying := v.Elem().Type().Name()
	if !strings.HasPrefix(underlying, "MapHandle[") {
		panic(fmt.Sprintf("fakeFTL received *%s, expected *MapHandle", underlying))
	}
	return v.Pointer()
}

func actuallyCallMap(ctx context.Context, impl mapImpl) any {
	out, err := impl(ctx)
	if err != nil {
		panic(err)
	}
	return out
}

func (f *fakeFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any) error {
	f.publishWaitGroup.Add(1)
	f.globalTopic.PublishSync(publishEvent{topic: topic, content: event})
	return nil
}

// addSubscriber adds a subscriber to the fake FTL instance. Each subscriber included in the test must be manually added
func addSubscriber[E any](f *fakeFTL, sub ftl.SubscriptionHandle[E], sink ftl.Sink[E]) {
	f.pubSubLock.Lock()
	defer f.pubSubLock.Unlock()

	if _, ok := f.subscriptions[sub.Name]; !ok {
		f.subscriptions[sub.Name] = &subscription{
			name:   sub.Name,
			topic:  sub.Topic,
			errors: map[int]error{},
		}
	}

	subscribers, ok := f.subscribers[sub.Name]
	if !ok {
		subscribers = []subscriber{}
	}
	f.subscribers[sub.Name] = append(subscribers, func(ctx context.Context, event any) error {
		if event, ok := event.(E); ok {
			return sink(ctx, event)
		}
		return fmt.Errorf("unexpected event type %T for subscription %s", event, sub.Name)
	})
}

// eventsForTopic returns all events published to a topic
func eventsForTopic[E any](ctx context.Context, f *fakeFTL, topic ftl.TopicHandle[E]) []E {
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
func resultsForSubscription[E any](ctx context.Context, f *fakeFTL, handle ftl.SubscriptionHandle[E]) []SubscriptionResult[E] {
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
	for i := 0; i < count; i++ {
		e := topic[i]
		if event, ok := e.(E); ok {
			results = append(results, SubscriptionResult[E]{
				Event: event,
				Error: subscription.errors[i],
			})
		} else {
			logger.Warnf("unexpected event type %T for subscription %s", e, handle.Name)
		}

	}
	return results
}

func (f *fakeFTL) watchPubSub(ctx context.Context) {
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

func (f *fakeFTL) handlePubSubEvent(ctx context.Context, e pubSubEvent) {
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
func (f *fakeFTL) waitForSubscriptionsToComplete(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("pubsub")
	startTime := time.Now()
	nextLogTime := startTime.Add(2 * time.Second)

	// Make sure all published events make it into our pubsub state
	f.publishWaitGroup.Wait()

	for {
		if func() bool {
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
			if time.Now().After(nextLogTime) {
				// print out what we are waiting on
				nextLogTime = time.Now().Add(2 * time.Second)
				logger.Infof("waiting on subscriptions to complete after %ds:\n%s", int(time.Until(startTime).Seconds()*-1), strings.Join(slices.Map(remaining, func(r remainingState) string {
					var suffix string
					if r.isExecuting {
						suffix = ", 1 executing"
					}
					return fmt.Sprintf("  %s: %d events pending%s", r.name, r.pendingEvents, suffix)
				}), "\n"))
			}
			return false
		}() {
			// reached idle state
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}
