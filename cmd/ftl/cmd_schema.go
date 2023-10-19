package main

import (
	"context"
	"fmt"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/golang/protobuf/proto"
	"github.com/otiai10/copy"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type schemaCmd struct {
	Get      getSchemaCmd      `default:"" cmd:"" help:"Retrieve the cluster FTL schema."`
	Protobuf schemaProtobufCmd `cmd:"" help:"Generate protobuf schema mirroring the FTL schema structure."`
	Generate schemaGenerateCmd `cmd:"" help:"Stream the schema from the cluster and generate files from the template."`
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

type schemaGenerateCmd struct {
	Template string `arg:"" help:"Template directory to use." type:"existingdir"`
	Dest     string `arg:"" help:"Destination directory to write files to (will be erased)."`
}

func (s *schemaGenerateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	logger := log.FromContext(ctx)
	modules := map[string]*schema.Module{}
	regenerate := false
	for stream.Receive() {
		msg := stream.Msg()
		switch msg.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
			module, err := schema.ModuleFromProto(msg.Schema)
			if err != nil {
				return errors.Wrap(err, "invalid module schema")
			}
			modules[module.Name] = module

		case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
			delete(modules, msg.ModuleName)
		}
		if !msg.More {
			regenerate = true
		}
		if !regenerate {
			continue
		}
		if err := os.RemoveAll(s.Dest); err != nil {
			return errors.WithStack(err)
		}
		for _, module := range modules {
			if err := copy.Copy(s.Template, s.Dest); err != nil {
				return errors.WithStack(err)
			}
			if err := internal.Scaffold(s.Dest, module); err != nil {
				return errors.WithStack(err)
			}
		}
		logger.Infof("Generated %d modules in %s", len(modules), s.Dest)
	}
	return nil
}
