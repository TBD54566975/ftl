//ftl:module two
package two

import "context"

type User struct {
	Name string
}

type Request struct{}

type Response struct{}

func Two(ctx context.Context, req Request) (Response, error) {
	return Response{}, nil
}
