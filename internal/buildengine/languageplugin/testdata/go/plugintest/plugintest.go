package plugintest

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
	// uncommentForDependency: "ftl/dependable"
)

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Verbaabbcc(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	// uncommentForDependency: fmt.Printf("printing %v", dependable.Data{})
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
