// Code generated by FTL. DO NOT EDIT.
package subscriber

import (
	"context"
	ftlpubsub "ftl/pubsub"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type ConsumesSubscriptionFromExternalTopicClient func(context.Context, ftlpubsub.PayinEvent) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			ConsumesSubscriptionFromExternalTopic,
		),
	)
}