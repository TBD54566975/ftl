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

var _ = ftl.Subscription(publisher.TestTopic, "testTopicSubscription")
var _ = ftl.Subscription(publisher.Topic2, "doomedSubscription")
var _ = ftl.Subscription(publisher.Topic2, "doomedSubscription2")

var catchCount atomic.Value[int]

//ftl:verb
//ftl:subscribe testTopicSubscription
func Consume(ctx context.Context, req publisher.PubSubEvent) error {
	ftl.LoggerFromContext(ctx).Infof("Subscriber is consuming %v", req.Time)
	return nil
}

//ftl:verb
//ftl:subscribe doomedSubscription
//ftl:retry 2 1s 1s catch catch
func ConsumeButFailAndRetry(ctx context.Context, req publisher.PubSubEvent) error {
	return fmt.Errorf("always error: event %v", req.Time)
}

//ftl:verb
//ftl:subscribe doomedSubscription2
//ftl:retry 1 1s 1s catch catchAny
func ConsumeButFailAndCatchAny(ctx context.Context, req publisher.PubSubEvent) error {
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

//ftl:verb
func CatchAny(ctx context.Context, req builtin.CatchRequest[any]) error {
	if req.Verb.Module != "subscriber" {
		return fmt.Errorf("unexpected verb module: %v", req.Verb.Module)
	}
	if req.Verb.Name != "consumeButFailAndCatchAny" {
		return fmt.Errorf("unexpected verb name: %v", req.Verb.Name)
	}
	if req.RequestType != "publisher.PubSubEvent" {
		return fmt.Errorf("unexpected request type: %v", req.RequestType)
	}
	requestMap, ok := req.Request.(map[string]any)
	if !ok {
		return fmt.Errorf("expected request to be a map[string]any: %T", req.Request)
	}
	timeValue, ok := requestMap["time"]
	if !ok {
		return fmt.Errorf("expected request to have a time key")
	}
	if _, ok := timeValue.(string); !ok {
		return fmt.Errorf("expected request to have a time value of type string")
	}
	return nil
}
