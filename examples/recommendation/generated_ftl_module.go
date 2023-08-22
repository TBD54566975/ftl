//ftl:module recommendation
package recommendation

import (
  "context"
)

type ListRequest struct {
  UserID string `json:"userID"`
  UserProductIDs []string `json:"userProductIDs"`
}

type ListResponse struct {
  ProductIDs []string `json:"productIDs"`
}


//ftl:verb
func List(context.Context, ListRequest) (ListResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
