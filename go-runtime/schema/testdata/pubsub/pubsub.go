package pubsub

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type PayinEvent struct {
	Name string
}

type Payins = ftl.TopicHandle[PayinEvent]

//ftl:verb
func Payin(ctx context.Context, topic Payins) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Test"}); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

type PaymentProcessing = ftl.SubscriptionHandle[Payins, ProcessPayinClient, PayinEvent]

//ftl:verb
func ProcessPayin(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received PubSub event: %v", event)
	return nil
}

// publicBroadcast is a topic that broadcasts payin events to the public.
// out of order with subscription registration to test ordering doesn't matter.
//
//ftl:export
type PublicBroadcast = ftl.TopicHandle[PayinEvent]

//ftl:verb export
func Broadcast(ctx context.Context, topic PublicBroadcast) error {
	if err := topic.Publish(ctx, PayinEvent{Name: "Broadcast"}); err != nil {
		return fmt.Errorf("failed to publish broadcast event: %w", err)
	}
	return nil
}

type BroadcastSubscription = ftl.SubscriptionHandle[PublicBroadcast, ProcessBroadcastClient, PayinEvent]

//ftl:verb
//ftl:retry 10 1s
func ProcessBroadcast(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received broadcast event: %v", event)
	return nil
}
