//ftl:module time
package time

import (
  "context"
  "time"
)

type TimeRequest struct {
}

type TimeResponse struct {
  Time time.Time `json:"time"`
}

// Time returns the current time.
//
//ftl:verb
func Time(context.Context, TimeRequest) (TimeResponse, error) {
  panic("Verb stubs should not be called directly, instead use github.com/TBD54566975/ftl/runtime-go/sdk.Call()")
}
