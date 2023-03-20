package sdkgo

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/common/socket"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type moduleKey struct{}

func ContextWithModule(ctx context.Context, module string) context.Context {
	return context.WithValue(ctx, moduleKey{}, module)
}

func ModuleFromContext(ctx context.Context) string {
	value := ctx.Value(moduleKey{})
	if value == nil {
		panic("no module in context")
	}
	return value.(string) //nolint:forcetypeassert
}

type clientKey struct{}

// ContextWithClient returns a context with an ftlv1.VerbServiceClient attached.
func ContextWithClient(ctx context.Context, endpoint socket.Socket) (context.Context, error) {
	conn, err := socket.DialGRPC(ctx, endpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	client := ftlv1.NewVerbServiceClient(conn)
	return context.WithValue(ctx, clientKey{}, client), nil
}

// ClientFromContext returns the ftlv1.VerbServiceClient from the context, or panics.
func ClientFromContext(ctx context.Context) ftlv1.VerbServiceClient {
	value := ctx.Value(clientKey{})
	if value == nil {
		panic("no ftlv1.VerbServiceClient in context")
	}
	return value.(ftlv1.VerbServiceClient) //nolint:forcetypeassert
}
