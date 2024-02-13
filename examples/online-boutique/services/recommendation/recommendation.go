//ftl:module recommendation
package recommendation

import (
	"context"
	"fmt"
	"math/rand"

	"ftl/builtin"
	"ftl/productcatalog"

	ftl "github.com/TBD54566975/ftl/go-runtime/ftl"
)

type ListRequest struct {
	UserID         string
	UserProductIDs []string
}

type ListResponse struct {
	ProductIDs []string
}

//ftl:verb
//ftl:ingress GET /recommendation
func List(ctx context.Context, req builtin.HttpRequest[ListRequest]) (builtin.HttpResponse[ListResponse], error) {
	cresp, err := ftl.Call(ctx, productcatalog.List, builtin.HttpRequest[productcatalog.ListRequest]{})
	if err != nil {
		return builtin.HttpResponse[ListResponse]{Body: ListResponse{}}, fmt.Errorf("%s: %w", "failed to retrieve product catalog", err)
	}

	// Remove user-provided products from the catalog, to avoid recommending
	// them.
	userIDs := make(map[string]struct{}, len(req.Body.UserProductIDs))
	for _, id := range req.Body.UserProductIDs {
		userIDs[id] = struct{}{}
	}
	filtered := make([]string, 0, len(cresp.Body.Products))
	for _, product := range cresp.Body.Products {
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
	return builtin.HttpResponse[ListResponse]{Body: ListResponse{ProductIDs: ret}}, nil

}
