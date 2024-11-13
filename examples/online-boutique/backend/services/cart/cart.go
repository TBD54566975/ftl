//ftl:module cart
package cart

import (
	"context"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
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

//ftl:ingress POST /cart/add
func AddItem(ctx context.Context, req builtin.HttpRequest[AddItemRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[AddItemResponse, ftl.Unit], error) {
	store.Add(req.Body.UserID, req.Body.Item)
	return builtin.HttpResponse[AddItemResponse, ftl.Unit]{
		Body: ftl.Some(AddItemResponse{}),
	}, nil
}

type GetCartRequest struct {
	UserID string `json:"userID"`
}

//ftl:ingress GET /cart
func GetCart(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, GetCartRequest]) (builtin.HttpResponse[Cart, ftl.Unit], error) {
	return builtin.HttpResponse[Cart, ftl.Unit]{
		Body: ftl.Some(Cart{Items: store.Get(req.Query.UserID), UserID: req.Query.UserID}),
	}, nil
}

type EmptyCartRequest struct {
	UserID string `json:"userID"`
}

type EmptyCartResponse struct{}

//ftl:ingress POST /cart/empty
func EmptyCart(ctx context.Context, req builtin.HttpRequest[EmptyCartRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[EmptyCartResponse, ftl.Unit], error) {
	store.Empty(req.Body.UserID)
	return builtin.HttpResponse[EmptyCartResponse, ftl.Unit]{
		Body: ftl.Some(EmptyCartResponse{}),
	}, nil
}
