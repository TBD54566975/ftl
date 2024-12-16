// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"ftl/time"

	"github.com/block/ftl/go-runtime/ftl"
)

type Default = ftl.Config[string]

// An echo request.
type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req EchoRequest, tc time.TimeClient, defaultName Default) (EchoResponse, error) {
	tresp, err := tc(ctx, time.TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}

	return EchoResponse{Message: fmt.Sprintf("Hello, %s!!! It is %s!", req.Name.Default(defaultName.Get(ctx)), tresp.Time)}, nil
}
