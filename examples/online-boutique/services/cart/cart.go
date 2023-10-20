//ftl:module cart
package cart

import (
	"context"
)

var store = NewStore()

type Item struct {
	ProductID string `json:"productID"`
	Quantity  int    `json:"quantity"`
}

type AddItemRequest struct {
	UserID string `json:"userID"`
	Item   Item   `json:"item"`
}

type AddItemResponse struct{}

type Cart struct {
	UserID string `json:"userID"`
	Items  []Item `json:"items"`
}

//ftl:verb
//ftl:ingress POST /cart/add
func AddItem(ctx context.Context, req AddItemRequest) (AddItemResponse, error) {
	store.Add(req.UserID, req.Item)
	return AddItemResponse{}, nil
}

type GetCartRequest struct {
	UserID string `json:"userID"`
}

//ftl:verb
//ftl:ingress GET /cart
func GetCart(ctx context.Context, req GetCartRequest) (Cart, error) {
	return Cart{Items: store.Get(req.UserID), UserID: req.UserID}, nil
}

type EmptyCartRequest struct {
	UserID string `json:"userID"`
}

type EmptyCartResponse struct{}

//ftl:verb
//ftl:ingress POST /cart/empty
func EmptyCart(ctx context.Context, req EmptyCartRequest) (EmptyCartResponse, error) {
	store.Empty(req.UserID)
	return EmptyCartResponse{}, nil
}
