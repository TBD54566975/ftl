package depcycle2

import (
	"context"
	"fmt"
	"ftl/depcycle1"
)

type Request struct{}
type Response struct {
	Message string
}

//ftl:verb export
func Cycle2(ctx context.Context, req Request) (Response, error) {
	var resp depcycle1.Response
	return Response{Message: fmt.Sprintf("cycle2 %s", resp)}, nil
}
