package buildengine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/pubsub"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

type BuildResult struct {
	Name   string
	Errors []*schema.Error
	Schema *schema.Module
}

// PluginEvent is used to notify of updates from the plugin.
//
//sumtype:decl
type PluginEvent interface {
	pluginEvent()
}

// AutoRebuildStartedEvent is sent when the plugin starts an automatic rebuild.
type AutoRebuildStartedEvent struct {
	Module string
}

func (AutoRebuildStartedEvent) pluginEvent() {}

// AutoRebuildEndedEvent is sent when the plugin ends an automatic rebuild.
type AutoRebuildEndedEvent struct {
	Module string
	Result either.Either[BuildResult, error]
}

func (AutoRebuildEndedEvent) pluginEvent() {}

// LanguagePlugin handles building and scaffolding modules in a specific language.
type LanguagePlugin interface {
	// Topic for all update events from the plugin
	// The same topic must be returned each time this method is called
	Updates() *pubsub.Topic[PluginEvent]

	// CreateModule creates a new module in the given directory with the given name and language.
	// Replacements and groups are special cases until plugins can provide their parameters.
	CreateModule(ctx context.Context, config moduleconfig.AbsModuleConfig, includeBinDir bool, replacements map[string]string, group string) error

	// GetDependencies returns the dependencies of the module.
	GetDependencies(ctx context.Context) ([]string, error)

	// Build builds the module with the latest config and schema.
	// In dev mode, plugin is responsible for automatically rebuilding as relevant files within the module change,
	// and publishing these automatic builds updates to Updates().
	Build(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error)

	// TODO: docs
	Kill(ctx context.Context) error
}

// PluginFromConfig creates a new language plugin from the given config.
func PluginFromConfig(ctx context.Context, config moduleconfig.AbsModuleConfig, projectRoot string) (p LanguagePlugin, err error) {
	switch config.Language {
	case "go":
		return newGoPlugin(ctx, config), nil
	case "java", "kotlin":
		return newJavaPlugin(ctx, config), nil
	case "rust":
		return newRustPlugin(ctx, config), nil
	default:
		return p, fmt.Errorf("unknown language %q", config.Language)
	}
}

var scaffoldFuncs = template.FuncMap{
	"snake":          strcase.ToLowerSnake,
	"screamingSnake": strcase.ToUpperSnake,
	"camel":          strcase.ToUpperCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"strippedCamel":  strcase.ToUpperStrippedCamel,
	"kebab":          strcase.ToLowerKebab,
	"screamingKebab": strcase.ToUpperKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename":       schema.TypeName,
}

//sumtype:decl
type pluginCommand interface {
	pluginCmd()
}

type buildCommand struct {
	projectRoot string
	config      moduleconfig.AbsModuleConfig
	schema      *schema.Schema
	buildEnv    []string
	devMode     bool

	result chan either.Either[BuildResult, error]
}

func (buildCommand) pluginCmd() {}

type dependenciesFunc = func() ([]string, error)
type getDependenciesCommand struct {
	dependenciesFunc dependenciesFunc

	result chan either.Either[[]string, error]
}

func (getDependenciesCommand) pluginCmd() {}

type buildFunc = func(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction ModifyFilesTransaction) error

// internalPlugin is used by languages that have not been split off into their own external plugins yet.
// It has standard behaviours around building and watching files.
type internalPlugin struct {
	// config does not change, may not be up to date
	config moduleconfig.AbsModuleConfig

	// build is called when a new build is explicitly requested or when a watched file changes
	buildFunc buildFunc

	// commands to execute
	commands chan pluginCommand

	updates *pubsub.Topic[PluginEvent]
}

func newInternalPlugin(ctx context.Context, config moduleconfig.AbsModuleConfig, build buildFunc) *internalPlugin {
	plugin := &internalPlugin{
		config:    config,
		buildFunc: build,
		commands:  make(chan pluginCommand, 128),
		updates:   pubsub.New[PluginEvent](),
	}
	go plugin.run(ctx)
	return plugin
}

func (p *internalPlugin) build(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, schema *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error) {
	cmd := buildCommand{
		projectRoot: projectRoot,
		config:      config,
		schema:      schema,
		buildEnv:    buildEnv,
		devMode:     devMode,
		result:      make(chan either.Either[BuildResult, error]),
	}
	p.commands <- cmd
	select {
	case result := <-cmd.result:
		switch result := result.(type) {
		case either.Left[BuildResult, error]:
			return result.Get(), nil
		case either.Right[BuildResult, error]:
			return BuildResult{}, result.Get() //nolint:wrapcheck
		default:
			panic(fmt.Sprintf("unexpected result type %T", result))
		}
	case <-ctx.Done():
		return BuildResult{}, ctx.Err()
	}
}

func (p *internalPlugin) getDependencies(ctx context.Context, d dependenciesFunc) ([]string, error) {
	cmd := getDependenciesCommand{
		dependenciesFunc: d,
		result:           make(chan either.Either[[]string, error]),
	}
	p.commands <- cmd
	select {
	case result := <-cmd.result:
		switch result := result.(type) {
		case either.Left[[]string, error]:
			return result.Get(), nil
		case either.Right[[]string, error]:
			return nil, result.Get()
		default:
			panic(fmt.Sprintf("unexpected result type %T", result))
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *internalPlugin) run(ctx context.Context) {
	watcher := NewWatcher(p.config.Watch...)
	watchChan := make(chan WatchEvent, 128)

	// state
	config := p.config
	var projectRoot string
	var schema *schema.Schema
	var buildEnv []string
	devMode := false

	for {
		select {
		case cmd := <-p.commands:
			switch c := cmd.(type) {
			case buildCommand:
				// update state
				projectRoot = c.projectRoot
				config = c.config
				schema = c.schema
				buildEnv = c.buildEnv
				if c.devMode && !devMode {
					devMode = true
					topic, err := watcher.Watch(ctx, time.Second, []string{config.Dir})
					if err != nil {
						c.result <- either.RightOf[BuildResult](fmt.Errorf("failed to start watching: %w", err))
						continue
					}
					topic.Subscribe(watchChan)
				}

				// build
				transaction := watcher.GetTransaction(p.config.Dir)
				err := p.buildFunc(ctx, projectRoot, config, schema, buildEnv, devMode, transaction)
				if err != nil {
					c.result <- either.RightOf[BuildResult](err)
					continue
				}

				result, err := loadBuildResult(config)
				if err != nil {
					c.result <- either.RightOf[BuildResult](err)
					continue
				}
				c.result <- either.LeftOf[error](result)

			case getDependenciesCommand:
				result, err := c.dependenciesFunc()
				if err != nil {
					c.result <- either.RightOf[[]string](err)
					continue
				}
				c.result <- either.LeftOf[error](result)
			}
		case event := <-watchChan:
			switch event.(type) {
			case WatchEventModuleChanged:
				// automatic rebuild
				p.updates.Publish(AutoRebuildStartedEvent{Module: p.config.Module})
				transaction := watcher.GetTransaction(p.config.Dir)
				err := p.buildFunc(ctx, projectRoot, config, schema, buildEnv, devMode, transaction)
				if err != nil {
					p.updates.Publish(AutoRebuildEndedEvent{
						Module: p.config.Module,
						Result: either.RightOf[BuildResult](err),
					})
					continue
				}
				result, err := loadBuildResult(config)
				if err != nil {
					p.updates.Publish(AutoRebuildEndedEvent{
						Module: p.config.Module,
						Result: either.RightOf[BuildResult](err),
					})
					continue
				}
				p.updates.Publish(AutoRebuildEndedEvent{
					Module: p.config.Module,
					Result: either.LeftOf[error](result),
				})
			case WatchEventModuleAdded:
				// ignore

			case WatchEventModuleRemoved:
				// ignore
			}

		case <-ctx.Done():
			return
		}
	}
}

// loadBuildResult reads the result of a build (ie schema and errors) from disk.
// internal plugins don't have a way to pass back schema and errors other than writing them to disk.
func loadBuildResult(config moduleconfig.AbsModuleConfig) (BuildResult, error) {
	errorList, err := loadProtoErrors(config)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read build errors for module: %w", err)
	}

	result := BuildResult{
		Errors: errorList.Errors,
	}

	if schema.ContainsTerminalError(errorList.Errors) {
		// skip reading schema
		return result, nil
	}

	sch, err := schema.ModuleFromProtoFile(config.Schema())
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read schema for module: %w", err)
	}
	result.Schema = sch
	return result, nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*schema.ErrorList, error) {
	if _, err := os.Stat(config.Errors); errors.Is(err, os.ErrNotExist) {
		return &schema.ErrorList{Errors: make([]*schema.Error, 0)}, nil
	}

	content, err := os.ReadFile(config.Errors)
	if err != nil {
		return nil, err
	}
	errorspb := &schemapb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, err
	}
	return schema.ErrorListFromProto(errorspb), nil
}
