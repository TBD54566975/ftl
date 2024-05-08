package omitempty

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct{}

type EchoResponse struct {
	Error   string `json:"error,omitempty"` // Should be omitted from marshaled JSON
	MustSet string `json:"mustset"`         // Should marshal to `"mustset":""`
}

//ftl:ingress http GET /get
func Get(ctx context.Context, req builtin.HttpRequest[EchoRequest]) (builtin.HttpResponse[EchoResponse, string], error) {
	return EchoResponse{}, nil
}
