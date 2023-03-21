package metadata

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const directRoutingKey = "ftl.direct"

// WithDirectRouting ensures any hops in Verb routing do not redirect.
//
// This is used so that eg. calls from Drives do not create recursive loops
// when calling back to the Agent.
func WithDirectRouting(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, directRoutingKey, "1")
}

// IsDirectRouted returns true if the incoming request should be directly
// routed and never redirected.
func IsDirectRouted(ctx context.Context) bool {
	return len(metadata.ValueFromIncomingContext(ctx, directRoutingKey)) > 0
}

// MetadataUnaryServerInterceptor propagates FTL metadata through to client calls.
func MetadataUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = propagateMetadata(ctx, md)
		}
		return handler(ctx, req)
	}
}

// MetadataStreamServerInterceptor propagates FTL metadata through to client calls.
func MetadataStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = propagateMetadata(ss.Context(), md)
		}
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: ctx})
	}
}

func propagateMetadata(ctx context.Context, md metadata.MD) context.Context {
	out := metadata.MD{}
	for key, values := range md {
		if strings.HasPrefix(key, "ftl.") {
			out[key] = values
		}
	}
	return metadata.NewOutgoingContext(ctx, out)
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ctx *wrappedServerStream) Context() context.Context { return ctx.ctx }
