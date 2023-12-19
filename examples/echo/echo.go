// This is the echo module.
//
//ftl:module echo
package echo

import (
	"context"
	"fmt"

	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/sdk"
)

// An echo request.
type EchoRequest struct {
	Name sdk.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// Echo returns a greeting with the current time.
//
//ftl:verb
//ftl:ingress GET /echo
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	fmt.Println("Echo received a request!")
	tresp, err := sdk.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!!! It is %s!", req.Name.Default("anonymous"), tresp.Time)}, nil
}
