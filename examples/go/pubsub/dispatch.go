package pubsub

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
	"golang.org/x/exp/rand"
)

type OrderPizzaRequest struct {
	Type     ftl.Option[string] `json:"type"`
	Customer string             `json:"customer"`
}

type OrderPizzaResponse struct {
	ID int `json:"id"`
}

//ftl:verb export
func OrderPizza(ctx context.Context, req OrderPizzaRequest, topic NewOrderTopic) (OrderPizzaResponse, error) {
	randomID := rand.Intn(1000)
	p := Pizza{
		ID:       randomID,
		Type:     req.Type.Default("cheese"),
		Customer: req.Customer,
	}
	ftl.LoggerFromContext(ctx).Infof("Ordering pizza with ID: %d", randomID)
	topic.Publish(ctx, p)
	return OrderPizzaResponse{ID: randomID}, nil
}

//ftl:verb
//ftl:subscribe pubsub.pizzaReadyTopic from=beginning
func DeliverPizza(ctx context.Context, pizza Pizza) error {
	ftl.LoggerFromContext(ctx).Infof("Delivering pizza: %v", pizza)
	return nil
}
