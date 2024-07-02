package integer

import (
	"context"
)

type EchoRequest struct {
	Value int64 `json:"value"`
}

type EchoResponse struct {
	Value int64 `json:"value"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Value: req.Value}, nil
}
