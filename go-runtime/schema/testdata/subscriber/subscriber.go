package subscriber

import (
	"context"
	"ftl/pubsub"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type SubscriptionToExternalTopic = ftl.SubscriptionHandle[pubsub.PublicBroadcast, ConsumesSubscriptionFromExternalTopicClient, pubsub.PayinEvent]

//ftl:verb
func ConsumesSubscriptionFromExternalTopic(ctx context.Context, req pubsub.PayinEvent) error {
	return nil
}
