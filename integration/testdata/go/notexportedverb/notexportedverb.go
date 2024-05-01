package notexportedverb

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Request struct {
	Name ftl.Option[string] `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

//ftl:verb
func ShouldFail(ctx context.Context, req Request) (Response, error) {
	_, err := ftl.Call(ctx, echo.Echo, echo.EchoRequest{})
	if err != nil {
		return Response{}, err
	}
	return Response{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
