package notexportedverb

import (
	"context"
	"fmt"

	"ftl/echo"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Request struct {
	Name ftl.Option[string] `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

//ftl:verb
func ShouldFail(ctx context.Context, req Request, ec echo.EchoClient) (Response, error) {
	_, err := ec(ctx, echo.EchoRequest{})
	if err != nil {
		return Response{}, err
	}
	return Response{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
