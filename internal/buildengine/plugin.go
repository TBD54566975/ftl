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

//sumtype:decl
type PluginEvent interface {
	pluginEvent()
}

type AutoRebuildStartedEvent struct {
	Module string
	// TODO: include file change for logging?
}

func (AutoRebuildStartedEvent) pluginEvent() {}

type AutoRebuildEndedEvent struct {
	Module string
	Result either.Either[BuildResult, error]
}

func (AutoRebuildEndedEvent) pluginEvent() {}

// TODO: docs
type Plugin interface {
	// Topic for all update events from the plugin
	Updates() *pubsub.Topic[PluginEvent]

	// CreateModule creates a new module in the given directory with the given name and language.
	CreateModule(ctx context.Context, config moduleconfig.AbsModuleConfig) error
	// TODO: docs
	GetDependencies(ctx context.Context) ([]string, error)
	// TODO: docs
	Build(ctx context.Context, config moduleconfig.AbsModuleConfig, sch *schema.Schema, projectPath string, buildEnv []string, devMode bool) (BuildResult, error)

	// TODO: docs
	Kill(ctx context.Context) error
}

func PluginFromConfig(ctx context.Context, config moduleconfig.AbsModuleConfig, projectPath string) (p Plugin, err error) {
	switch config.Language {
	case "go":
		return newGoPlugin(ctx, config, projectPath), nil
	case "java", "kotlin":
		return newJavaPlugin(ctx, config, projectPath), nil
	case "rust":
		return newRustPlugin(ctx, config, projectPath), nil
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

type pluginCommand interface {
	pluginCmd()
}

type buildCommand struct {
	config   moduleconfig.AbsModuleConfig
	schema   *schema.Schema
	buildEnv []string
	devMode  bool

	// TODO: turn this into an either[BuildResult, error] channel?
	result chan either.Either[BuildResult, error]
}

func (buildCommand) pluginCmd() {}

type watchCommand struct {
	err chan error
}

func (watchCommand) pluginCmd() {}

type dependenciesFunc = func() ([]string, error)
type getDependenciesCommand struct {
	dependenciesFunc dependenciesFunc

	result chan either.Either[[]string, error]
}

func (getDependenciesCommand) pluginCmd() {}

type buildFunc = func(ctx context.Context, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction ModifyFilesTransaction) error

// internal plugin is used by languages that have not been split off into their own external plugins yet.
// it coordinates builds and watching for changes
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

func (p *internalPlugin) build(ctx context.Context, config moduleconfig.AbsModuleConfig, schema *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error) {
	if devMode {
		watchCmd := watchCommand{
			err: make(chan error),
		}
		p.commands <- watchCmd
		err := <-watchCmd.err
		if err != nil {
			return BuildResult{}, err
		}
	}
	cmd := buildCommand{
		config:   config,
		schema:   schema,
		buildEnv: buildEnv,
		devMode:  devMode,
		result:   make(chan either.Either[BuildResult, error]),
	}
	p.commands <- cmd
	select {
	case result := <-cmd.result:
		switch result := result.(type) {
		case either.Left[BuildResult, error]:
			return result.Get(), nil
		case either.Right[BuildResult, error]:
			return BuildResult{}, result.Get()
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
	isWatching := false
	config := p.config
	var schema *schema.Schema
	var buildEnv []string
	devMode := false

	for {
		select {
		case cmd := <-p.commands:
			// TODO: commands can be queued up and file changes can be ignored if there is also a build command in the queue
			switch c := cmd.(type) {
			case buildCommand:
				// update state
				config = c.config
				schema = c.schema
				buildEnv = c.buildEnv
				if c.devMode {
					devMode = true
				}

				transaction := watcher.GetTransaction(p.config.Dir)
				err := p.buildFunc(ctx, config, schema, buildEnv, devMode, transaction)
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
			case watchCommand:
				if isWatching {
					c.err <- nil
					continue
				}
				isWatching = true
				topic, err := watcher.Watch(ctx, time.Second, []string{config.Dir})
				if err != nil {
					c.err <- fmt.Errorf("failed to start watching: %w", err)
					continue
				}
				topic.Subscribe(watchChan)
				c.err <- nil
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
				// build
				p.updates.Publish(AutoRebuildStartedEvent{Module: p.config.Module})
				transaction := watcher.GetTransaction(p.config.Dir)
				err := p.buildFunc(ctx, config, schema, buildEnv, devMode, transaction)
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
