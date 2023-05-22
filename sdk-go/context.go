package sdkgo

import (
	"context"

	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/socket"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type clientKey struct{}

// ContextWithClient returns a context with an ftlv1.VerbServiceClient attached.
func ContextWithClient(ctx context.Context, endpoint socket.Socket) context.Context {
	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint.URL().String())
	return context.WithValue(ctx, clientKey{}, client)
}

// ClientFromContext returns the ftlv1.VerbServiceClient from the context, or panics.
func ClientFromContext(ctx context.Context) ftlv1connect.VerbServiceClient {
	value := ctx.Value(clientKey{})
	if value == nil {
		panic("no ftlv1connect.VerbServiceClient in context")
	}
	return value.(ftlv1connect.VerbServiceClient) //nolint:forcetypeassert
}
