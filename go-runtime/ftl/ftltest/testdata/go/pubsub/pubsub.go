package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
var Topic = ftl.Topic[Event]("topic")

//ftl:export
var Topic2 = ftl.Topic[Event]("topic2")

var subscription = ftl.Subscription(Topic, "subscription")

//ftl:data
type Event struct {
	Value string
}

//ftl:verb
func PublishToTopicOne(ctx context.Context, event Event) error {
	return Topic.Publish(ctx, event)
}

//ftl:verb
func PublishToTopicTwo(ctx context.Context, event Event) error {
	return Topic2.Publish(ctx, event)
}

//ftl:subscribe subscription
func ErrorsAfterASecond(ctx context.Context, event Event) error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("SubscriberThatFails always fails")
}
