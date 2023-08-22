//ftl:module productcatalog
package productcatalog

import (
  "context"
)

type ListRequest struct {
}

type Money struct {
  CurrencyCode string `json:"currencyCode"`
  Units int `json:"units"`
  Nanos int `json:"nanos"`
}

type Product struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Description string `json:"description"`
  Picture string `json:"picture"`
  PriceUSD Money `json:"priceUSD"`
  Categories []string `json:"categories"`
}

type ListResponse struct {
  Products []Product `json:"products"`
}


//ftl:verb
func List(context.Context, ListRequest) (ListResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}

type GetRequest struct {
  Id string `json:"id"`
}


//ftl:verb
func Get(context.Context, GetRequest) (Product, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}

type SearchRequest struct {
  Query string `json:"query"`
}

type SearchResponse struct {
  Results []Product `json:"results"`
}


//ftl:verb
func Search(context.Context, SearchRequest) (SearchResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
