package pubsub

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:export
var NewOrderTopic = ftl.Topic[Pizza]("newOrderTopic")

//ftl:export
var PizzaReadyTopic = ftl.Topic[Pizza]("pizzaReadyTopic")

var _ = ftl.Subscription(NewOrderTopic, "cookPizzaSub")

type Pizza struct {
	ID       int
	Type     string
	Customer string
}

//ftl:subscribe cookPizzaSub
func CookPizza(ctx context.Context, pizza Pizza) error {
	ftl.LoggerFromContext(ctx).Infof("Cooking pizza: %v", pizza)
	return PizzaReadyTopic.Publish(ctx, pizza)
}
