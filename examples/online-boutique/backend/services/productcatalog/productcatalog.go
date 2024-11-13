//ftl:module productcatalog
package productcatalog

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"ftl/currency"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/examples/online-boutique/backend/common"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var (
	//go:embed database.json
	databaseJSON []byte
	database     = common.LoadDatabase[[]Product](databaseJSON)
)

type Product struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Picture     string         `json:"picture"`
	PriceUSD    currency.Money `json:"priceUsd"`

	// Categories such as "clothing" or "kitchen" that can be used to look up
	// other related products.
	Categories []string `json:"categories"`
}

type ListRequest struct{}

type ListResponse struct {
	Products []Product `json:"products"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

//ftl:ingress GET /productcatalog
func List(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ListRequest]) (builtin.HttpResponse[ListResponse, ftl.Unit], error) {
	return builtin.HttpResponse[ListResponse, ftl.Unit]{
		Body: ftl.Some(ListResponse{Products: database}),
	}, nil
}

type GetRequest struct {
	ID string
}

//ftl:ingress GET /productcatalog/{id}
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetRequest, ftl.Unit]) (builtin.HttpResponse[Product, ErrorResponse], error) {
	for _, p := range database {
		if p.ID == req.PathParameters.ID {
			return builtin.HttpResponse[Product, ErrorResponse]{Body: ftl.Some(p)}, nil
		}
	}
	return builtin.HttpResponse[Product, ErrorResponse]{Error: ftl.Some(ErrorResponse{Message: fmt.Sprintf("product not found: %q", req.PathParameters.ID)})}, nil
}

type SearchRequest struct {
	Query string
}

type SearchResponse struct {
	Results []Product
}

//ftl:verb
func Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	out := SearchResponse{}
	q := strings.ToLower(req.Query)
	for _, p := range database {
		if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Description), q) {
			out.Results = append(out.Results, p)
		}
	}
	return out, nil
}
