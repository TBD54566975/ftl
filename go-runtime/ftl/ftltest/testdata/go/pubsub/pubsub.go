package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

//ftl:export
type Topic1 = ftl.TopicHandle[Event, ftl.SinglePartitionMap[Event]]

//ftl:export
type Topic2 = ftl.TopicHandle[Event, ftl.SinglePartitionMap[Event]]

//ftl:data
type Event struct {
	Value string
}

//ftl:verb
func PublishToTopicOne(ctx context.Context, event Event, topic1 Topic1) error {
	return topic1.Publish(ctx, event)
}

//ftl:verb
//ftl:subscribe topic1 from=beginning
func PropagateToTopic2(ctx context.Context, event Event, topic2 Topic2) error {
	return topic2.Publish(ctx, event)
}

//ftl:verb
//ftl:subscribe topic2 from=beginning
func ConsumeEvent(_ context.Context, _ Event) error {
	return nil
}

//ftl:verb
//ftl:subscribe topic1 from=beginning
func ErrorsAfterASecond(ctx context.Context, event Event) error {
	time.Sleep(1 * time.Second)
	return fmt.Errorf("SubscriberThatFails always fails")
}
