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

//ftl:export
var Topic2 = ftl.Topic[Event]("topic_2")

var subscription1_1 = ftl.Subscription(Topic1, "subscription_1_1")
var subscription1_2 = ftl.Subscription(Topic1, "subscription_1_2")
var subscription2_1 = ftl.Subscription(Topic2, "subscription_2_1")
var subscription2_2 = ftl.Subscription(Topic2, "subscription_2_3")

//ftl:data
type Event struct {
	Value string
}

//ftl:verb
func PublishToTopicOne(ctx context.Context, event Event) error {
	return Topic1.Publish(ctx, event)
}

//ftl:verb
func PropagateToTopic2(ctx context.Context, event Event) error {
	return Topic2.Publish(ctx, event)
}

//ftl:verb
func ConsumeEvent(_ context.Context, _ Event) error {
	return nil
}

//ftl:subscribe subscription_1_1
func ErrorsAfterASecond(ctx context.Context, event Event) error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("SubscriberThatFails always fails")
}
