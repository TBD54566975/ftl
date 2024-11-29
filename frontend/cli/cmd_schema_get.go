package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/internal/schema"
)

type getSchemaCmd struct {
	Watch    bool     `help:"Watch for changes to the schema."`
	Protobuf bool     `help:"Output the schema as binary protobuf." xor:"format"`
	JSON     bool     `help:"Output the schema as JSON." xor:"format"`
	Modules  []string `help:"Modules to include" type:"string" optional:""`
}

func (g *getSchemaCmd) Run(ctx context.Context, client ftlv1connect.SchemaServiceClient) error {
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return err
	}
	if g.Protobuf {
		return g.generateProto(resp)
	}
	if g.JSON {
		return g.generateJSON(resp)
	}
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	for resp.Receive() {
		msg := resp.Msg()
		switch msg.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED:
			if msg.Schema == nil {
				return fmt.Errorf("schema is nil for added/changed deployment %q", msg.GetDeploymentKey())
			}
			module, err := schema.ModuleFromProto(msg.Schema)
			if len(g.Modules) == 0 || remainingNames[msg.Schema.Name] {
				if err != nil {
					return fmt.Errorf("invalid module schema: %w", err)
				}
				fmt.Println(module)
				delete(remainingNames, msg.Schema.Name)
			}
			if !msg.More {
				missingNames := maps.Keys(remainingNames)
				slices.Sort(missingNames)
				if len(missingNames) > 0 {
					if g.Watch {
						fmt.Printf("missing modules: %s\n", strings.Join(missingNames, ", "))
					} else {
						return fmt.Errorf("missing modules: %s", strings.Join(missingNames, ", "))
					}
				}
				if !g.Watch {
					return nil
				}
			}
		case ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED:
			if msg.Schema == nil {
				return fmt.Errorf("schema is nil for removed deployment %q", msg.GetDeploymentKey())
			}
			if msg.ModuleRemoved {
				delete(remainingNames, msg.Schema.Name)
			}
			fmt.Printf("deployment %s removed\n", msg.GetDeploymentKey())
		case ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED:
			return fmt.Errorf("unexpected unspecified deployment change type for deployment %q", msg.GetDeploymentKey())
		}

	}
	if err := resp.Err(); err != nil {
		return resp.Err()
	}
	return nil
}

func (g *getSchemaCmd) generateJSON(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	schema := &schemapb.Schema{}
	for resp.Receive() {
		msg := resp.Msg()
		if len(g.Modules) == 0 || remainingNames[msg.Schema.Name] {
			schema.Modules = append(schema.Modules, msg.Schema)
			delete(remainingNames, msg.Schema.Name)
		}
		if !msg.More {
			break
		}
	}
	if err := resp.Err(); err != nil {
		return fmt.Errorf("error receiving schema: %w", err)
	}
	data, err := protojson.Marshal(schema)
	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}
	fmt.Printf("%s\n", data)
	missingNames := maps.Keys(remainingNames)
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return fmt.Errorf("missing modules: %v", missingNames)
	}
	return nil
}

func (g *getSchemaCmd) generateProto(resp *connect.ServerStreamForClient[ftlv1.PullSchemaResponse]) error {
	remainingNames := make(map[string]bool)
	for _, name := range g.Modules {
		remainingNames[name] = true
	}
	schema := &schemapb.Schema{}
	for resp.Receive() {
		msg := resp.Msg()
		if len(g.Modules) == 0 || remainingNames[msg.Schema.Name] {
			schema.Modules = append(schema.Modules, msg.Schema)
			delete(remainingNames, msg.Schema.Name)
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
	if err != nil {
		return err
	}
	missingNames := maps.Keys(remainingNames)
	slices.Sort(missingNames)
	if len(missingNames) > 0 {
		return fmt.Errorf("missing modules: %v", missingNames)
	}
	return nil
}
