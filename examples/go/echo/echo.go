// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var defaultName = ftl.Config[string]("default")

// An echo request.
type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
	Age  ftl.Encrypted[int] `json:"age"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	tresp, err := ftl.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}

	return EchoResponse{Message: fmt.Sprintf("Hello, %s!!! It is %s .. age %d!",
		req.Name.Default(defaultName.Get(ctx)), tresp.Time),
		req.Age.Decrypt(ctx),
	}, nil
}
