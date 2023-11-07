package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/golang/protobuf/proto"
	"github.com/otiai10/copy"
	"github.com/radovskyb/watcher"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

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
	Watch    time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	Template string        `arg:"" help:"Template directory to use." type:"existingdir"`
	Dest     string        `arg:"" help:"Destination directory to write files to (will be erased)."`
}

func (s *schemaGenerateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	watch := watcher.New()
	defer watch.Close()

	absTemplatePath, err := filepath.Abs(s.Template)
	if err != nil {
		return errors.WithStack(err)
	}
	absDestPath, err := filepath.Abs(s.Dest)
	if err != nil {
		return errors.WithStack(err)
	}

	if strings.HasPrefix(absDestPath, absTemplatePath) {
		return fmt.Errorf("destination directory %s must not be inside the template directory %s", absDestPath, absTemplatePath)
	}

	logger := log.FromContext(ctx)
	logger.Infof("Watching %s", s.Template)

	if err := watch.AddRecursive(s.Template); err != nil {
		return errors.WithStack(err)
	}

	stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
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

			moduleChange <- maps.Values(modules)
		}
		return nil
	})

	wg.Go(func() error { return errors.WithStack(watch.Start(s.Watch)) })

	var previousModules []*schema.Module
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(wg.Wait())

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
	return nil
}
