package typeregistry

import (
	"context"
	"ftl/builtin"
	"ftl/typeregistry/subpackage"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct {
	Strings subpackage.StringsTypeEnum
}

type EchoResponse struct {
	Strings subpackage.StringsTypeEnum
}

//ftl:ingress POST /echo
func Echo(ctx context.Context, req builtin.HttpRequest[EchoRequest]) (builtin.HttpResponse[EchoResponse, string], error) {
	return builtin.HttpResponse[EchoResponse, string]{
		Body: ftl.Some(EchoResponse{Strings: req.Body.Strings}),
	}, nil
}
