package typeregistry

import (
	"context"

	"ftl/builtin"

	"ftl/typeregistry/subpackage"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type EchoRequest struct {
	Strings subpackage.StringsTypeEnum
}

type EchoResponse struct {
	Strings subpackage.StringsTypeEnum
}

//ftl:ingress POST /echo
func Echo(ctx context.Context, req builtin.HttpRequest[EchoRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[EchoResponse, string], error) {
	return builtin.HttpResponse[EchoResponse, string]{
		Body: ftl.Some(EchoResponse{Strings: req.Body.Strings}),
	}, nil
}
