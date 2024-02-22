package buildengine

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/rpc"
)

func buildGo(ctx context.Context, module Module) error {
	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	resp, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return err
	}
	sch, err := schema.FromProto(resp.Msg.Schema)
	if err != nil {
		return fmt.Errorf("failed to convert schema from proto: %w", err)
	}
	if err := compile.Build(ctx, module.Dir, sch); err != nil {
		return fmt.Errorf("failed to build module: %w", err)
	}
	return nil
}
