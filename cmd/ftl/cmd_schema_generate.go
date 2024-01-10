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
	"github.com/TBD54566975/scaffolder/extensions/javascript"
	"github.com/radovskyb/watcher"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type schemaGenerateCmd struct {
	Watch    time.Duration `short:"w" help:"Watch template directory at this frequency and regenerate on change."`
	Template string        `arg:"" help:"Template directory to use." type:"existingdir"`
	Dest     string        `arg:"" help:"Destination directory to write files to (will be erased)."`
}

func (s *schemaGenerateCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	if s.Watch == 0 {
		return s.oneOffGenerate(ctx, client)
	}
	return s.hotReload(ctx, client)
}

func (s *schemaGenerateCmd) oneOffGenerate(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	response, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}
	modules, err := slices.MapErr(response.Msg.Schema.Modules, schema.ModuleFromProto)
	if err != nil {
		return fmt.Errorf("invalid module schema: %w", err)
	}
	return s.regenerateModules(log.FromContext(ctx), modules)
}

func (s *schemaGenerateCmd) hotReload(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
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
		if err := scaffolder.Scaffold(s.Template, s.Dest, module,
			scaffolder.Functions(scaffoldFuncs),
			scaffolder.Extend(javascript.Extension("template.js", javascript.WithLogger(makeJSLoggerAdapter(logger)))),
		); err != nil {
			return err
		}
	}
	logger.Infof("Generated %d modules in %s", len(modules), s.Dest)
	return nil
}

func makeJSLoggerAdapter(logger *log.Logger) func(args ...any) {
	return func(args ...any) {
		strs := slices.Map(args, func(v any) string { return fmt.Sprintf("%v", v) })
		level := log.Info
		if prefix, ok := args[0].(string); ok {
			switch prefix {
			case "log:":
				level = log.Info
			case "debug:":
				level = log.Debug
			case "error:":
				level = log.Error
			case "warn:":
				level = log.Warn
			}
		}
		logger.Log(log.Entry{
			Level:   level,
			Message: strings.Join(strs[1:], " "),
		})
	}
}
