package depcycle1

import (
	"context"
	"fmt"
	"ftl/depcycle2"
)

type Request struct{}
type Response struct {
	Message string
}

//ftl:verb export
func Cycle1(ctx context.Context, req Request) (Response, error) {
	var resp depcycle2.Response
	return Response{Message: fmt.Sprintf("cycle1 %s", resp)}, nil
}
