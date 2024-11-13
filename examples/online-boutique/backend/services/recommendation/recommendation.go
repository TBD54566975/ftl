//ftl:module recommendation
package recommendation

import (
	"context"
	"fmt"
	"math/rand"

	"ftl/productcatalog"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type ListRequest struct {
	UserID         string
	UserProductIDs []string
}

type ListResponse struct {
	ProductIDs []string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

//ftl:ingress GET /recommendation
func List(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ListRequest], client productcatalog.ListClient) (builtin.HttpResponse[ListResponse, ErrorResponse], error) {
	cresp, err := client(ctx, builtin.HttpRequest[ftl.Unit, ftl.Unit, productcatalog.ListRequest]{})
	if err != nil {
		return builtin.HttpResponse[ListResponse, ErrorResponse]{
			Error: ftl.Some(ErrorResponse{Message: fmt.Sprintf("%s: %w", "failed to retrieve product catalog", err)}),
		}, nil
	}

	listResponse, ok := cresp.Body.Get()
	if !ok {
		return builtin.HttpResponse[ListResponse, ErrorResponse]{
			Error: ftl.Some(ErrorResponse{Message: "failed to retrieve product catalog"}),
		}, nil
	}

	// Remove user-provided products from the catalog, to avoid recommending
	// them.
	userIDs := make(map[string]struct{}, len(req.Query.UserProductIDs))
	for _, id := range req.Query.UserProductIDs {
		userIDs[id] = struct{}{}
	}
	filtered := make([]string, 0, len(listResponse.Products))
	for _, product := range listResponse.Products {
		if _, ok := userIDs[product.Id]; ok {
			continue
		}
		filtered = append(filtered, product.Id)
	}

	// Sample from filtered products and return them.
	perm := rand.Perm(len(filtered))
	const maxResponses = 5
	ret := make([]string, 0, maxResponses)
	for _, idx := range perm {
		ret = append(ret, filtered[idx])
		if len(ret) >= maxResponses {
			break
		}
	}
	return builtin.HttpResponse[ListResponse, ErrorResponse]{Body: ftl.Some(ListResponse{ProductIDs: ret})}, nil

}
