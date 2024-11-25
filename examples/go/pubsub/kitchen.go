package pubsub

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:export
type NewOrderTopic = ftl.TopicHandle[Pizza]

//ftl:export
type PizzaReadyTopic = ftl.TopicHandle[Pizza]

type CookPizzaSub = ftl.SubscriptionHandle[NewOrderTopic, CookPizzaClient, Pizza]

type Pizza struct {
	ID       int
	Type     string
	Customer string
}

//ftl:verb
func CookPizza(ctx context.Context, pizza Pizza, topic PizzaReadyTopic) error {
	ftl.LoggerFromContext(ctx).Infof("Cooking pizza: %v", pizza)
	return topic.Publish(ctx, pizza)
}
