package validateconfig

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Default = ftl.Config[string]
type Count = ftl.Config[int]

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest, defaultName Default) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default(defaultName.Get(ctx)))}, nil
}
