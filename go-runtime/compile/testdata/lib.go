package lib

import "context"

type Request struct {
}

type Response struct {
}

func OtherFunc(ctx context.Context, req Request) (Response, error) {
	return Response{}, nil
}

type NonFTLType struct {
	Message string
}
