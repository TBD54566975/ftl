package subscriber

import (
	"context"
	"fmt"
	"ftl/publisher"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var _ = ftl.Subscription(publisher.TestTopic, "test_subscription")

//ftl:verb
//ftl:subscribe test_subscription
func Consume(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Subscriber is consuming %v", req.Time)
	return nil
}

var _ = ftl.Subscription(publisher.Topic2, "doomed_subscription")

//ftl:verb
//ftl:subscribe doomed_subscription
//ftl:retry 2 1s 1s
func ConsumeButFailAndRetry(ctx context.Context, req publisher.PubSubEvent) error {
	return fmt.Errorf("always error: event %v", req.Time)
}

//ftl:verb
func PublishToExternalModule(ctx context.Context) error {
	// Get around compile-time checks
	var topic = publisher.TestTopic
	return topic.Publish(ctx, publisher.PubSubEvent{Time: time.Now()})
}
