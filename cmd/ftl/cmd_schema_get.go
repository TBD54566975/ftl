package main

import (
	"context"
	"fmt"
	"os"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
)

type getSchemaCmd struct {
	Protobuf bool `help:"Output the schema as binary protobuf."`
}

func (g *getSchemaCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return err
	}
	if g.Protobuf {
		return g.generateProto(resp)
	}
	for resp.Receive() {
		msg := resp.Msg()
		module, err := schema.ModuleFromProto(msg.Schema)
		if err != nil {
			return fmt.Errorf("%s: %w", "invalid module schema", err)
		}
		fmt.Println(module)
		if !msg.More {
			break
		}
	}
	return resp.Err()
}

func (g *getSchemaCmd) generateProto(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	schema := &schemapb.Schema{}
	for resp.Receive() {
		msg := resp.Msg()
		schema.Modules = append(schema.Modules, msg.Schema)
		if !msg.More {
			break
		}
	}
	if err := resp.Err(); err != nil {
		return err
	}
	pb, err := proto.Marshal(schema)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(pb)
	return err
}
