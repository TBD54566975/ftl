package rpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

func TestRPCContext(t *testing.T) {
	ctx := context.Background()
	verbClient := ftlv1connect.NewVerbServiceClient(http.DefaultClient, "http://localhost:8080")
	ctx = ContextWithClient(ctx, verbClient)
	controlplaneClient := ftlv1connect.NewControlPlaneServiceClient(http.DefaultClient, "http://localhost:8080")
	ctx = ContextWithClient(ctx, controlplaneClient)

	assert.Equal(t, verbClient, ClientFromContext[ftlv1connect.VerbServiceClient](ctx))
	assert.Equal(t, controlplaneClient, ClientFromContext[ftlv1connect.ControlPlaneServiceClient](ctx))
}
