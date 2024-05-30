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
	if err := payinsVar.Publish(PayinEvent{Name: "Test"}); err != nil {
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

var _ = ftl.RegisterSubscription(payinsVar, "paymentProcessing")

var payinsVar = ftl.RegisterTopic[PayinEvent]("payins")

var _ = ftl.RegisterSubscription(broadcast, "broadcastSubscription")

// publicBroadcast is a topic that broadcasts payin events to the public.
// out of order with subscription registration to test ordering doesn't matter.
//
//ftl:export
var broadcast = ftl.RegisterTopic[PayinEvent]("publicBroadcast")

//ftl:verb
func Broadcast(ctx context.Context) error {
	if err := broadcast.Publish(PayinEvent{Name: "Broadcast"}); err != nil {
		return fmt.Errorf("failed to publish broadcast event: %w", err)
	}
	return nil
}

//ftl:subscribe broadcastSubscription
func ProcessBroadcast(ctx context.Context, event PayinEvent) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Received broadcast event: %v", event)
	return nil
}
