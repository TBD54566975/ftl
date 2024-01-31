//ftl:module echo
package echo

import (
	"context"
	"fmt"

	"ftl/echo2"
	ftl "github.com/TBD54566975/ftl/go-runtime/sdk" // Import the FTL SDK.
)

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}

//ftl:verb
func Call(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	res, err := ftl.Call(ctx, echo2.Echo, echo2.EchoRequest{Name: req.Name})
	if err != nil {
		return EchoResponse{}, err
	}
	return EchoResponse{Message: res.Message}, nil
}
