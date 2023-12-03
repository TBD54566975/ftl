package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/scaffolder"
	"github.com/golang/protobuf/proto"
	"github.com/radovskyb/watcher"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/schema"
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

type schemaGenerateCmd struct {
	Watch    time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	Template string        `arg:"" help:"Template directory to use." type:"existingdir"`
	Dest     string        `arg:"" help:"Destination directory to write files to (will be erased)."`
}

func (s *schemaGenerateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	watch := watcher.New()
	defer watch.Close()

	absTemplatePath, err := filepath.Abs(s.Template)
	if err != nil {
		return err
	}
	absDestPath, err := filepath.Abs(s.Dest)
	if err != nil {
		return err
	}

	if strings.HasPrefix(absDestPath, absTemplatePath) {
		return fmt.Errorf("destination directory %s must not be inside the template directory %s", absDestPath, absTemplatePath)
	}

	logger := log.FromContext(ctx)
	logger.Infof("Watching %s", s.Template)

	if err := watch.AddRecursive(s.Template); err != nil {
		return err
	}

	stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return err
	}
	wg, ctx := errgroup.WithContext(ctx)

	moduleChange := make(chan []*schema.Module)

	wg.Go(func() error {
		modules := map[string]*schema.Module{}
		regenerate := false
		for stream.Receive() {
			msg := stream.Msg()
			switch msg.ChangeType {
			case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
				module, err := schema.ModuleFromProto(msg.Schema)
				if err != nil {
					return fmt.Errorf("%s: %w", "invalid module schema", err)
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

			moduleChange <- maps.Values(modules)
		}
		return nil
	})

	wg.Go(func() error { return watch.Start(s.Watch) })

	var previousModules []*schema.Module
	for {
		select {
		case <-ctx.Done():
			return wg.Wait()

		case event := <-watch.Event:
			logger.Infof("Template changed (%s), regenerating modules", event.Path)
			if err := s.regenerateModules(logger, previousModules); err != nil {
				return err
			}

		case modules := <-moduleChange:
			previousModules = modules
			if err := s.regenerateModules(logger, modules); err != nil {
				return err
			}
		}
	}
}

func (s *schemaGenerateCmd) regenerateModules(logger *log.Logger, modules []*schema.Module) error {
	if err := os.RemoveAll(s.Dest); err != nil {
		return err
	}
	for _, module := range modules {
		if err := scaffolder.Scaffold(s.Template, s.Dest, module, scaffolder.Functions(scaffoldFuncs)); err != nil {
			return err
		}
	}
	logger.Infof("Generated %d modules in %s", len(modules), s.Dest)
	return nil
}
