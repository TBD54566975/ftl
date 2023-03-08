package echo

import (
	"context"
	"fmt"
)

type GoodbyeRequest struct{ Name string }
type GoodbyeResponse struct{ Message string }

//ftl:verb
func Goodbye(ctx context.Context, req GoodbyeRequest) (GoodbyeResponse, error) {
	return GoodbyeResponse{Message: fmt.Sprintf("Goodbye, %s!", req.Name)}, nil
}
