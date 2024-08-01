package console

import (
	"context"
	"fmt"

	"ftl/builtin"

	lib "github.com/TBD54566975/ftl/backend/controller/console/testdata/go"
	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type External = lib.External

type Response struct {
	Message string
}

//ftl:ingress http GET /test
func Get(ctx context.Context, req builtin.HttpRequest[External]) (builtin.HttpResponse[Response, string], error) {
	return builtin.HttpResponse[Response, string]{
		Body: ftl.Some(Response{
			Message: fmt.Sprintf("Hello, %s", req.Body.Message),
		}),
	}, nil
}

//ftl:verb
func Verb(ctx context.Context, req External) (External, error) {
	return External{
		Message: fmt.Sprintf("Hello, %s", req.Message),
	}, nil
}
