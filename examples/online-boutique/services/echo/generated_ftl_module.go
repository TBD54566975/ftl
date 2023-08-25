//ftl:module echo
package echo

import (
  "context"
)

type EchoRequest struct {
  Name string `json:"name"`
}

type EchoResponse struct {
  Message string `json:"message"`
}

// Echo returns a greeting with the current time.
//
//ftl:verb
func Echo(context.Context, EchoRequest) (EchoResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
