//ftl:module ad
package ad

import (
  "context"
)

type AdRequest struct {
  ContextKeys []string `json:"contextKeys"`
}

type Ad struct {
  RedirectURL string `json:"redirectURL"`
  Text string `json:"text"`
}

type AdResponse struct {
  Ads []Ad `json:"ads"`
}


//ftl:verb
func Get(context.Context, AdRequest) (AdResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
