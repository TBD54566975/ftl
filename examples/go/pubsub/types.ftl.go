// Code generated by FTL. DO NOT EDIT.
package pubsub

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type CookPizzaClient func(context.Context, Pizza) error

type DeliverPizzaClient func(context.Context, Pizza) error

type OrderPizzaClient func(context.Context, OrderPizzaRequest) (OrderPizzaResponse, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			CookPizza,
		),
		reflection.ProvideResourcesForVerb(
			DeliverPizza,
		),
		reflection.ProvideResourcesForVerb(
			OrderPizza,
		),
	)
}