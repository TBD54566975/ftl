package pubsub

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type PayinEvent struct {
	Name string
}

//ftl:verb
func Payin(ctx context.Context) error {
	if err := Payins.Publish(ctx, PayinEvent{Name: "Test"}); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

//ftl:subscribe paymentProcessing
func ProcessPayin(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received PubSub event: %v", event)
	return nil
}

var _ = ftl.Subscription(Payins, "paymentProcessing")

var Payins = ftl.Topic[PayinEvent]("payins")

var _ = ftl.Subscription(PublicBroadcast, "broadcastSubscription")

// publicBroadcast is a topic that broadcasts payin events to the public.
// out of order with subscription registration to test ordering doesn't matter.
//
//ftl:export
var PublicBroadcast = ftl.Topic[PayinEvent]("publicBroadcast")

//ftl:verb export
func Broadcast(ctx context.Context) error {
	if err := PublicBroadcast.Publish(ctx, PayinEvent{Name: "Broadcast"}); err != nil {
		return fmt.Errorf("failed to publish broadcast event: %w", err)
	}
	return nil
}

//ftl:subscribe broadcastSubscription
//ftl:retry 10 1s
func ProcessBroadcast(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received broadcast event: %v", event)
	return nil
}
