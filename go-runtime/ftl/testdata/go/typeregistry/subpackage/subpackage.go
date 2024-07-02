package subpackage

import (
	"context"
	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:enum
type StringsTypeEnum interface {
	tag()
}

type Single string

func (Single) tag() {}

type List []string

func (List) tag() {}

type Object struct {
	S string
}

func (Object) tag() {}

type EchoRequest struct {
	Strings StringsTypeEnum
}

type EchoResponse struct {
	Strings StringsTypeEnum
}

//ftl:ingress POST /echo
func Echo(ctx context.Context, req builtin.HttpRequest[EchoRequest]) (builtin.HttpResponse[EchoResponse, string], error) {
	return builtin.HttpResponse[EchoResponse, string]{
		Body: ftl.Some(EchoResponse{Strings: req.Body.Strings}),
	}, nil
}
