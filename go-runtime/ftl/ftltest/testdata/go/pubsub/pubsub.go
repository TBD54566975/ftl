package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
type Topic1 = ftl.TopicHandle[Event]

//ftl:export
type Topic2 = ftl.TopicHandle[Event]

type Subscription1_1_ErrorsAfterASecond = ftl.SubscriptionHandle[Topic1, ErrorsAfterASecondClient, Event]
type Subscription1_1_PropagateToTopic2 = ftl.SubscriptionHandle[Topic1, PropagateToTopic2Client, Event]
type Subscription1_2_PropagateToTopic2 = ftl.SubscriptionHandle[Topic1, PropagateToTopic2Client, Event]
type Subscription2_1_ConsumeEvent = ftl.SubscriptionHandle[Topic2, ConsumeEventClient, Event]
type Subscription2_2_ConsumeEvent = ftl.SubscriptionHandle[Topic2, ConsumeEventClient, Event]

//ftl:data
type Event struct {
	Value string
}

//ftl:verb
func PublishToTopicOne(ctx context.Context, event Event, topic1 Topic1) error {
	return topic1.Publish(ctx, event)
}

//ftl:verb
func PropagateToTopic2(ctx context.Context, event Event, topic2 Topic2) error {
	return topic2.Publish(ctx, event)
}

//ftl:verb
func ConsumeEvent(_ context.Context, _ Event) error {
	return nil
}

//ftl:verb
func ErrorsAfterASecond(ctx context.Context, event Event) error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("SubscriberThatFails always fails")
}
