// This is the echo module.
package echo

import (
	"context"
	"fmt"

	"ftl/builtin"
	"ftl/time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var defaultName = ftl.Config[string]("default")

type EchoEvent struct {
	Message string
}

var Echotopic = ftl.Topic[EchoEvent]("echotopic")
var sub = ftl.Subscription(Echotopic, "sub")

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
//ftl:ingress POST /echotopic
func Echo(ctx context.Context, req builtin.HttpRequest[EchoRequest]) (builtin.HttpResponse[EchoResponse, string], error) {
	tresp, err := ftl.Call(ctx, time.Time, time.TimeRequest{})
	if err != nil {
		return builtin.HttpResponse[EchoResponse, string]{}, err
	}
	if err := Echotopic.Publish(ctx, EchoEvent{Message: "adf"}); err != nil {
		return builtin.HttpResponse[EchoResponse, string]{}, err
	}

	return builtin.HttpResponse[EchoResponse, string]{Body: ftl.Some(EchoResponse{Message: fmt.Sprintf("Hello, %s!!! It is %s!", req.Body.Name.Default(defaultName.Get(ctx)), tresp.Time)})}, nil
}

//ftl:verb
//ftl:subscribe sub
func EchoSinkOne(ctx context.Context, e EchoEvent) error {
	return nil
}
