//ftl:module basket
package basket

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/sdk/kvstore"
)

//ftl:resource
var kv = kvstore.Require[Basket]()

type Basket struct {
	Items []string
}

type ItemRequest struct {
	BasketID string
	ItemID   string
}

type BasketSummary struct {
	Items int
}

type BasketRequest struct {
	ID string
}

// Add an item to the basket.
//
//ftl:verb
func Add(ctx context.Context, order ItemRequest) (BasketSummary, error) {
	panic("??")
}

// Remove an item from the basket.
//
//ftl:verb
func Remove(ctx context.Context, item ItemRequest) (BasketSummary, error) {
	panic("??")
}

func Get(ctx context.Context, basket BasketRequest) (Basket, error) {
	panic("??")
}
