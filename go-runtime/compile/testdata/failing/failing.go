package failing

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type Request struct{}
type Response struct{}

//ftl:export
func FailingVerb(ctx context.Context, req Request) (Response, error) {
	ftl.Call(ctx, "failing", "failingVerb", req)
	return Response{}, nil
}
