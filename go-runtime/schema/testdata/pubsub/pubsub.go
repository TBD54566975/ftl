package pubsub

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type PayinEvent struct {
	Name string
}

type Payins = ftl.TopicHandle[PayinEvent, ftl.SinglePartitionMap[PayinEvent]]

//ftl:verb
func Payin(ctx context.Context, topic Payins) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Test"}); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

//ftl:verb
//ftl:subscribe payins from=beginning
func ProcessPayin(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received PubSub event: %v", event)
	return nil
}

// publicBroadcast is a topic that broadcasts payin events to the public.
// out of order with subscription registration to test ordering doesn't matter.
//
//ftl:export
type PublicBroadcast = ftl.TopicHandle[PayinEvent, ftl.SinglePartitionMap[PayinEvent]]

//ftl:verb export
func Broadcast(ctx context.Context, topic PublicBroadcast) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Broadcast"}); err != nil {
		return fmt.Errorf("failed to publish broadcast event: %w", err)
	}
	return nil
}

//ftl:verb
//ftl:subscribe publicBroadcast from=beginning
//ftl:retry 10 1s
func ProcessBroadcast(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received broadcast event: %v", event)
	return nil
}
