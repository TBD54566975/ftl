package subscriber

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ftl/builtin"
	"ftl/publisher"

	"github.com/alecthomas/atomic"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var _ = ftl.Subscription(publisher.TestTopic, "testSubscription")

var catchCount atomic.Value[int]

//ftl:verb
//ftl:subscribe testSubscription
func Consume(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Subscriber is consuming %v", req.Time)
	return nil
}

var _ = ftl.Subscription(publisher.Topic2, "doomedSubscription")

//ftl:verb
//ftl:subscribe doomedSubscription
//ftl:retry 2 1s 1s catch catch
func ConsumeButFailAndRetry(ctx context.Context, req publisher.PubSubEvent) error {
	return fmt.Errorf("always error: event %v", req.Time)
}

//ftl:verb
func PublishToExternalModule(ctx context.Context) error {
	// Get around compile-time checks
	var topic = publisher.TestTopic
	return topic.Publish(ctx, publisher.PubSubEvent{Time: time.Now()})
}

//ftl:verb
func Catch(ctx context.Context, req builtin.CatchRequest[publisher.PubSubEvent]) error {
	if !strings.Contains(req.Error, "always error: event") {
		return fmt.Errorf("unexpected error: %v", req.Error)
	}
	count := catchCount.Load() + 1
	catchCount.Store(count)

	// fail once
	if count == 1 {
		return fmt.Errorf("catching error")
	}

	// succeed after that
	return nil
}
