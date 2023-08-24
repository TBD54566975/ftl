//ftl:module cart
package cart

import (
	"context"
)

var store = NewStore()

type Item struct {
	ProductID string
	Quantity  int
}

type AddItemRequest struct {
	UserID string
	Item   Item
}

type AddItemResponse struct{}

type Cart struct {
	UserID string
	Items  []Item
}

//ftl:verb
func AddItem(ctx context.Context, req AddItemRequest) (AddItemResponse, error) {
	store.Add(req.UserID, req.Item)
	return AddItemResponse{}, nil
}

type GetCartRequest struct {
	UserID string
}

//ftl:verb
func GetCart(ctx context.Context, req GetCartRequest) (Cart, error) {
	return Cart{Items: store.Get(req.UserID)}, nil
}

type EmptyCartRequest struct {
	UserID string
}

type EmptyCartResponse struct{}

//ftl:verb
func EmptyCart(ctx context.Context, req EmptyCartRequest) (EmptyCartResponse, error) {
	store.Empty(req.UserID)
	return EmptyCartResponse{}, nil
}
