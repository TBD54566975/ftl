package rpc

import (
	"context"
	"net/http"
	"testing"

	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/alecthomas/assert/v2"
)

func TestRPCContext(t *testing.T) {
	ctx := context.Background()
	verbClient := ftlv1connect.NewVerbServiceClient(http.DefaultClient, "http://localhost:8080")
	ctx = ContextWithClient(ctx, verbClient)
	controllerClient := ftlv1connect.NewControllerServiceClient(http.DefaultClient, "http://localhost:8080")
	ctx = ContextWithClient(ctx, controllerClient)

	assert.Equal(t, verbClient, ClientFromContext[ftlv1connect.VerbServiceClient](ctx))
	assert.Equal(t, controllerClient, ClientFromContext[ftlv1connect.ControllerServiceClient](ctx))
}
