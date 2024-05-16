// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var defaultName = ftl.Config[string]("default")

//ftl:enum
type Enum interface{ enum() }

type A int

func (A) enum() {}

type B string

func (B) enum() {}

// An echo request.
type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// Echo returns a greeting with the current time.
//
//ftl:verb
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	tresp, err := ftl.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}

	return EchoResponse{Message: fmt.Sprintf("Hello, %s!!! It is %s!", req.Name.Default(defaultName.Get(ctx)), tresp.Time)}, nil
}
