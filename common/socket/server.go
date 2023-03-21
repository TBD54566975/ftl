package socket

import (
	"context"

	"google.golang.org/grpc"

	"github.com/TBD54566975/ftl/common/metadata"
)

// NewGRPCServer returns a new gRPC server with values from the context merged
// in, and FTL metadata propagated from incoming server methods to outbound client calls.
func NewGRPCServer(ctx context.Context) *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(contextUnaryInterceptor(ctx), metadata.MetadataUnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(contextStreamInterceptor(ctx), metadata.MetadataStreamServerInterceptor()),
	)
}

func contextUnaryInterceptor(root context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = mergedContext{values: root, Context: ctx}
		return handler(ctx, req)
	}
}

func contextStreamInterceptor(root context.Context) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := mergedContext{values: root, Context: ss.Context()}
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

var _ context.Context = (*mergedContext)(nil)

type mergedContext struct {
	values context.Context
	context.Context
}

func (m mergedContext) Value(key any) any {
	if value := m.Context.Value(key); value != nil {
		return value
	}
	return m.values.Value(key)
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ctx *wrappedServerStream) Context() context.Context { return ctx.ctx }
