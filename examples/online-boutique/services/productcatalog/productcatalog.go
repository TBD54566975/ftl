//ftl:module productcatalog
package productcatalog

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/TBD54566975/ftl/examples/online-boutique/common"
	"github.com/TBD54566975/ftl/examples/online-boutique/common/money"
)

var (
	//go:embed database.json
	databaseJSON []byte
	database     = common.LoadDatabase[[]Product](databaseJSON)
)

type Product struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Picture     string      `json:"picture"`
	PriceUSD    money.Money `json:"priceUSD"`

	// Categories such as "clothing" or "kitchen" that can be used to look up
	// other related products.
	Categories []string `json:"categories"`
}

type ListRequest struct{}

type ListResponse struct {
	Products []Product `json:"products"`
}

//ftl:verb
//ftl:ingress GET /productcatalog
func List(ctx context.Context, req ListRequest) (ListResponse, error) {
	return ListResponse{Products: database}, nil
}

type GetRequest struct {
	ID string
}

//ftl:verb
//ftl:ingress GET /productcatalog/id
func Get(ctx context.Context, req GetRequest) (Product, error) {
	for _, p := range database {
		if p.ID == req.ID {
			return p, nil
		}
	}
	return Product{}, fmt.Errorf("product not found: %q", req.ID)
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
