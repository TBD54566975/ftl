//ftl:module echo
package echo

import (
	"context"
	"fmt"
	"time"

	ftl "github.com/TBD54566975/ftl/sdk-go"
)

type EchoRequest struct {
	Name string `json:"name"`
}
type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	tresp, err := ftl.Call(ctx, Time, TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}
	t := time.Unix(int64(tresp.Time), 0)
	return EchoResponse{Message: fmt.Sprintf("Hello, %s! It is %s!", req.Name, t)}, nil
}
