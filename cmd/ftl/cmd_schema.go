package main

import (
	"context"
	"fmt"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/golang/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type schemaCmd struct {
	Get      getSchemaCmd      `default:"" cmd:"" help:"Retrieve the cluster FTL schema."`
	Protobuf schemaProtobufCmd `cmd:"" help:"Generate protobuf schema mirroring the FTL schema structure."`
}

type schemaProtobufCmd struct{}

func (c *schemaProtobufCmd) Run() error { //nolint:unparam
	fmt.Println(schema.ProtobufSchema())
	return nil
}

type getSchemaCmd struct {
	Protobuf bool `help:"Output the schema as binary protobuf."`
}

func (g *getSchemaCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	if g.Protobuf {
		return g.generateProto(resp)
	}
	for resp.Receive() {
		msg := resp.Msg()
		module, err := schema.ModuleFromProto(msg.Schema)
		if err != nil {
			return errors.Wrap(err, "invalid module schema")
		}
		fmt.Println(module)
		if !msg.More {
			break
		}
	}
	return errors.WithStack(resp.Err())
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
		return errors.WithStack(err)
	}
	pb, err := proto.Marshal(schema)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = os.Stdout.Write(pb)
	return errors.WithStack(err)
}
