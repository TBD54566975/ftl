//ftl:module cart
package cart

import (
	"context"
	"ftl/builtin"
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
func AddItem(ctx context.Context, req builtin.HttpRequest[AddItemRequest]) (builtin.HttpResponse[AddItemResponse], error) {
	store.Add(req.Body.UserID, req.Body.Item)
	return builtin.HttpResponse[AddItemResponse]{
		Body: AddItemResponse{},
	}, nil
}

type GetCartRequest struct {
	UserID string `json:"userID"`
}

//ftl:verb
//ftl:ingress GET /cart
func GetCart(ctx context.Context, req builtin.HttpRequest[GetCartRequest]) (builtin.HttpResponse[Cart], error) {
	return builtin.HttpResponse[Cart]{
		Body: Cart{Items: store.Get(req.Body.UserID), UserID: req.Body.UserID},
	}, nil
}

type EmptyCartRequest struct {
	UserID string `json:"userID"`
}

type EmptyCartResponse struct{}

//ftl:verb
//ftl:ingress POST /cart/empty
func EmptyCart(ctx context.Context, req builtin.HttpRequest[EmptyCartRequest]) (builtin.HttpResponse[EmptyCartResponse], error) {
	store.Empty(req.Body.UserID)
	return builtin.HttpResponse[EmptyCartResponse]{
		Body: EmptyCartResponse{},
	}, nil
}
