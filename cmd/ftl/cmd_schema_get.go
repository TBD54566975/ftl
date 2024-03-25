package main

import (
	"context"
	"fmt"
	"os"
	"slices"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
)

type getSchemaCmd struct {
	Protobuf bool     `help:"Output the schema as binary protobuf."`
	Modules  []string `arg:"" help:"Modules to include" type:"string" optional:""`
}

func (g *getSchemaCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return err
	}
	if g.Protobuf {
		return g.generateProto(resp)
	}
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	for resp.Receive() {
		msg := resp.Msg()
		module, err := schema.ModuleFromProto(msg.Schema)
		if len(g.Modules) == 0 || remainingNames[msg.Schema.Name] {
			if err != nil {
				return fmt.Errorf("%s: %w", "invalid module schema", err)
			}
			fmt.Println(module)
			delete(remainingNames, msg.Schema.Name)
		}
		if !msg.More {
			break
		}
	}
	if err := resp.Err(); err != nil {
		return resp.Err()
	}
	missingNames := maps.Keys(remainingNames)
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return fmt.Errorf("missing modules: %v", missingNames)
	}
	return nil
}

func (g *getSchemaCmd) generateProto(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	filterMap := make(map[string]bool)
	for _, name := range g.Modules {
		filterMap[name] = true
	}
	schema := &schemapb.Schema{}
	for resp.Receive() {
		msg := resp.Msg()
		if len(g.Modules) == 0 || filterMap[msg.Schema.Name] {
			schema.Modules = append(schema.Modules, msg.Schema)
		}
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
