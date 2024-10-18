package pkg

import (
	"context"
	"ftl/another"
)

func Pkg(ec another.EchoClient) {
	ec(context.Background(), another.EchoRequest{})
}
