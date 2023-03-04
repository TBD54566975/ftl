package log

import (
	"context"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

// UnaryGRPCInterceptor returns a grpc.UnaryServerInterceptor that adds a logger to the context.
func UnaryGRPCInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = ContextWithLogger(ctx, logger)
		return handler(ctx, req)
	}
}

// StreamGRPCInterceptor returns a grpc.StreamServerInterceptor that adds a logger to the context.
func StreamGRPCInterceptor(logger *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ContextWithLogger(ss.Context(), logger)
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ctx *wrappedServerStream) Context() context.Context { return ctx.ctx }
