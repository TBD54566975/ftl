package lib

import (
	"context"
	"fmt"
	"ftl/alpha"
)

func CreateEchoResponse(ctx context.Context, req alpha.EchoRequest) alpha.EchoResponse {
	return alpha.EchoResponse{Message: fmt.Sprintf("Hello, %s!!!", req.Name)}, nil
}
