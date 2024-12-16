package dischema

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type DefaultUser = ftl.Config[string]

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest, defaultUser DefaultUser) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default(defaultUser.Get(ctx)))}, nil
}
