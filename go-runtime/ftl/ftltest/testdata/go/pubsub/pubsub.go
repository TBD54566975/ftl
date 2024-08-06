package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
var Topic1 = ftl.Topic[Event]("topic_1")
var subscription1 = ftl.Subscription(Topic1, "subscription1")

var Topic2 = ftl.Topic[Event]("topic_2")
var subscription2 = ftl.Subscription(Topic2, "subscription2")

//ftl:data
type Event struct {
	Value string
}

//ftl:verb
func PublishToTopicOne(ctx context.Context, event Event) error {
	return Topic1.Publish(ctx, event)
}

//ftl:subscribe subscription1
func ErrorsAfterASecond(ctx context.Context, event Event) error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("SubscriberThatFails always fails")
}

//ftl:verb
func PublishToTopicTwo(ctx context.Context, event Event) error {
	return Topic2.Publish(ctx, event)
}
