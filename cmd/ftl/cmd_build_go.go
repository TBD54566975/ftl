package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func (b *buildCmd) buildGo(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	resp, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return err
	}
	sch, err := schema.FromProto(resp.Msg.Schema)
	if err != nil {
		return err
	}
	if err := compile.Build(ctx, b.ModuleDir, sch); err != nil {
		return fmt.Errorf("failed to build module: %w", err)
	}

	return nil
}
