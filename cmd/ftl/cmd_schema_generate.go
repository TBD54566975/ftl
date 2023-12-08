package main

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/repr"
	"github.com/dop251/goja"
	"github.com/iancoleman/strcase"
	"github.com/radovskyb/watcher"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/slices"
	"github.com/TBD54566975/ftl/backend/schema"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/scaffolder"
)

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
		funcs, _, err := s.createJSVM(logger)
		if err != nil {
			return fmt.Errorf("failed to create JS VM: %w", err)
		}
		if err := scaffolder.Scaffold(s.Template, s.Dest, module, scaffolder.Functions(funcs)); err != nil {
			return err
		}
	}
	logger.Infof("Generated %d modules in %s", len(modules), s.Dest)
	return nil
}

// Create JS VM and populate it with functions that can be used in templates.
func (s *schemaGenerateCmd) createJSVM(logger *log.Logger) (funcs template.FuncMap, vm *goja.Runtime, err error) {
	vm = goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	if err := initConsole(vm, logger); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("repr", repr.String); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("snake", strcase.ToSnake); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("screamingSnake", strcase.ToScreamingSnake); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("camel", strcase.ToCamel); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("lowerCamel", strcase.ToLowerCamel); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("kebab", strcase.ToKebab); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("screamingKebab", strcase.ToScreamingKebab); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("upper", strings.ToUpper); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("lower", strings.ToLower); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("title", strings.Title); err != nil {
		return nil, nil, err
	}
	if err := vm.Set("typename", func(v any) string {
		return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
	}); err != nil {
		return nil, nil, err
	}

	scriptPath := filepath.Join(s.Template, "template.js")
	if script, err := os.ReadFile(scriptPath); err == nil {
		if _, err := vm.RunScript(scriptPath, string(script)); err != nil {
			return nil, nil, fmt.Errorf("failed to run %s: %w", scriptPath, err)
		}
	}

	funcs = template.FuncMap{}
	global := vm.GlobalObject()
	for _, key := range global.Keys() {
		attr := global.Get(key)
		value := attr.Export()
		typ := reflect.TypeOf(value)
		if typ.Kind() != reflect.Func {
			continue
		}

		// Go functions are exported as is, JS functions are wrapped in a go function that calls them.
		isJsFunc := typ.NumIn() == 1 && typ.In(0) == reflect.TypeOf(goja.FunctionCall{})
		if !isJsFunc {
			funcs[key] = value
			continue
		}

		fn, ok := goja.AssertFunction(attr)
		if !ok {
			continue
		}
		funcs[key] = func(args ...any) (any, error) {
			vmArgs := slices.Map(args, vm.ToValue)
			return fn(global, vmArgs...)
		}
	}
	return funcs, vm, nil
}

// TODO: change from values...string to values...any
func initConsole(vm *goja.Runtime, logger *log.Logger) error {
	console := vm.NewObject()
	if err := console.Set("log", func(values ...any) {
		strs := slices.Map(values, func(v any) string { return fmt.Sprintf("%v", v) })
		logger.Infof("%s", strings.Join(strs, " "))
	}); err != nil {
		return err
	}
	if err := console.Set("debug", func(values ...any) {
		strs := slices.Map(values, func(v any) string { return fmt.Sprintf("%v", v) })
		logger.Debugf("%s", strings.Join(strs, " "))
	}); err != nil {
		return err
	}
	if err := console.Set("error", func(values ...any) {
		strs := slices.Map(values, func(v any) string { return fmt.Sprintf("%v", v) })
		logger.Logf(log.Error, "%s", strings.Join(strs, " "))
	}); err != nil {
		return err
	}
	if err := console.Set("warn", func(values ...any) {
		strs := slices.Map(values, func(v any) string { return fmt.Sprintf("%v", v) })
		logger.Warnf("%s", strings.Join(strs, " "))
	}); err != nil {
		return err
	}
	err := vm.Set("console", console)
	if err != nil {
		return err
	}
	return nil
}
