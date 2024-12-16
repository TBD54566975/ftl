package alpha

import (
	"context"
	"fmt"

	"ftl/other"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest, oc other.EchoClient) (EchoResponse, error) {
	oc(ctx, other.EchoRequest{})
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
