// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"golang.org/x/exp/rand"
)

var defaultName = ftl.Config[string]("default")

// An echo request.
type EchoRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

type FlakyEchoRequest struct {
	RandRange ftl.Option[int] `json:"randrange"`
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

// FlakyEcho returns a greeting with the current time.
//
//ftl:verb
func FlakyEcho(ctx context.Context, req FlakyEchoRequest) (EchoResponse, error) {
	rr, ok := req.RandRange.Get()
	if !ok {
		return nil, fmt.Errorf("RandRange field is missing")
	}
	if rand.Intn(rr) == 0 {
		return nil, fmt.Errorf("Gamble lost!")
	}

	tresp, err := ftl.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return EchoResponse{}, err
	}

	return EchoResponse{Message: fmt.Sprintf("It is %s!", tresp.Time)}, nil
}
