package pkg

import (
	"context"
	"ftl/another"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

func Pkg() {
	ftl.Call(context.Background(), another.Echo, another.EchoRequest{})
}
