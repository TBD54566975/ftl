package echo

import (
	"context"
	"fmt"

	ftl "github.com/TBD54566975/ftl/sdk-go"
)

type EchoRequest struct{ Name string }
type EchoResponse struct{ Message string }

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	time, err := ftl.Call(ctx, Time, TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}
	return EchoResponse{Message: fmt.Sprintf("Hello, %s! It is %s!", req.Name, time.Time)}, nil
}
