package mapper

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Gettable struct{ s string }

func (g Gettable) Get(ctx context.Context) string {
	return g.s
}

var gettable = Gettable{"my_string"}

var m = ftl.Map(gettable, func(ctx context.Context, c string) (int, error) {
	return len(c), nil
})

type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
