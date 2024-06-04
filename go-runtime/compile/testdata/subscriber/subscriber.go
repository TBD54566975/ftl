package subscriber

import (
	"context"
	"ftl/pubsub"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var _ = ftl.Subscription(pubsub.PublicBroadcast, "subscriptionToExternalTopic")

//ftl:subscribe subscriptionToExternalTopic
func ConsumesSubscriptionFromExternalTopic(ctx context.Context, req pubsub.PayinEvent) error {
	return nil
}
