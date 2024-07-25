package integer

import (
	"context"
)

type EchoRequest struct {
	Input int64 `json:"value"`
}

type EchoResponse struct {
	Output int64 `json:"value"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Output: req.Input}, nil
}
