package socket

import (
	"context"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/metadata"
	"google.golang.org/grpc"
)

func NewGRPCServer(ctx context.Context) *grpc.Server {
	logger := log.FromContext(ctx)
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(log.UnaryGRPCInterceptor(logger), metadata.MetadatUnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(log.StreamGRPCInterceptor(logger)),
	)
}
