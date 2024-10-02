package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/pubsub"
	"google.golang.org/protobuf/proto"

	languagepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

const BuildLockTimeout = time.Minute

type BuildResult struct {
	Name      string
	Errors    []builderrors.Error
	Schema    *schema.Module
	StartTime time.Time
}

// PluginEvent is used to notify of updates from the plugin.
//
//sumtype:decl
type PluginEvent interface {
	ModuleName() string
	pluginEvent()
}

// AutoRebuildStartedEvent is sent when the plugin starts an automatic rebuild.
type AutoRebuildStartedEvent struct {
	Module string
}

func (AutoRebuildStartedEvent) pluginEvent()         {}
func (e AutoRebuildStartedEvent) ModuleName() string { return e.Module }

// AutoRebuildEndedEvent is sent when the plugin ends an automatic rebuild.
type AutoRebuildEndedEvent struct {
	Module string
	Result either.Either[BuildResult, error]
}

func (AutoRebuildEndedEvent) pluginEvent()         {}
func (e AutoRebuildEndedEvent) ModuleName() string { return e.Module }

// LanguagePlugin handles building and scaffolding modules in a specific language.
type LanguagePlugin interface {
	// Topic for all update events from the plugin
	// The same topic must be returned each time this method is called
	Updates() *pubsub.Topic[PluginEvent]

	// GetCreateModuleFlags returns the flags that can be used to create a module for this language.
	GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error)

	// CreateModule creates a new module in the given directory with the given name and language.
	CreateModule(ctx context.Context, projConfig projectconfig.Config, moduleConfig moduleconfig.ModuleConfig, flags map[string]string) error

	// GetDependencies returns the dependencies of the module.
	GetDependencies(ctx context.Context) ([]string, error)

	// Build builds the module with the latest config and schema.
	// In dev mode, plugin is responsible for automatically rebuilding as relevant files within the module change,
	// and publishing these automatic builds updates to Updates().
	Build(ctx context.Context, projectRoot string, config moduleconfig.ModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error)

	// Kill stops the plugin and cleans up any resources.
	Kill(ctx context.Context) error
}

// PluginFromConfig creates a new language plugin from the given config.
func PluginFromConfig(ctx context.Context, config moduleconfig.ModuleConfig, projectRoot string) (p LanguagePlugin, err error) {
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

//sumtype:decl
type pluginCommand interface {
	pluginCmd()
}

type buildCommand struct {
	projectRoot string
	config      moduleconfig.ModuleConfig
	schema      *schema.Schema
	buildEnv    []string
	devMode     bool

	result chan either.Either[BuildResult, error]
}

func (buildCommand) pluginCmd() {}

type dependenciesFunc func() ([]string, error)
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
	config moduleconfig.ModuleConfig

	// build is called when a new build is explicitly requested or when a watched file changes
	buildFunc buildFunc

	// commands to execute
	commands chan pluginCommand

	updates *pubsub.Topic[PluginEvent]
	cancel  context.CancelFunc
}

func newInternalPlugin(ctx context.Context, config moduleconfig.ModuleConfig, build buildFunc) *internalPlugin {
	plugin := &internalPlugin{
		config:    config,
		buildFunc: build,
		commands:  make(chan pluginCommand, 128),
		updates:   pubsub.New[PluginEvent](),
	}
	ctx, plugin.cancel = context.WithCancel(ctx)
	go plugin.run(ctx)
	return plugin
}

func (p *internalPlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.updates
}

func (p *internalPlugin) Kill(ctx context.Context) error {
	p.cancel()
	return nil
}

func (p *internalPlugin) Build(ctx context.Context, projectRoot string, config moduleconfig.ModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool) (BuildResult, error) {
	cmd := buildCommand{
		projectRoot: projectRoot,
		config:      config,
		schema:      sch,
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
		return BuildResult{}, fmt.Errorf("error waiting for build to complete: %w", ctx.Err())
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
			return nil, fmt.Errorf("could not get dependencies: %w", result.Get())
		default:
			panic(fmt.Sprintf("unexpected result type %T", result))
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("could not get dependencies: %w", ctx.Err())
	}
}

func (p *internalPlugin) run(ctx context.Context) {
	watcher := NewWatcher(p.config.Watch...)
	watchChan := make(chan WatchEvent, 128)

	// State
	// This is updated when given explicit build commands and used for automatic rebuilds
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

				// begin watching if needed
				if c.devMode && !devMode {
					devMode = true
					topic, err := watcher.Watch(ctx, time.Second, []string{config.Abs().Dir})
					if err != nil {
						c.result <- either.RightOf[BuildResult](fmt.Errorf("failed to start watching: %w", err))
						continue
					}
					topic.Subscribe(watchChan)
				}

				// build
				result, err := buildAndLoadResult(ctx, projectRoot, config, schema, buildEnv, devMode, watcher, p.buildFunc)
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
				result, err := buildAndLoadResult(ctx, projectRoot, config, schema, buildEnv, devMode, watcher, p.buildFunc)
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

func buildAndLoadResult(ctx context.Context, projectRoot string, c moduleconfig.ModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, watcher *Watcher, build buildFunc) (BuildResult, error) {
	config := c.Abs()
	release, err := flock.Acquire(ctx, filepath.Join(config.Dir, ".ftl.lock"), BuildLockTimeout)
	if err != nil {
		return BuildResult{}, fmt.Errorf("could not acquire build lock for %v: %w", config.Module, err)
	}
	defer release() //nolint:errcheck

	startTime := time.Now()

	if err := os.RemoveAll(config.DeployDir); err != nil {
		return BuildResult{}, fmt.Errorf("failed to clear deploy directory: %w", err)
	}
	if err := os.MkdirAll(config.DeployDir, 0700); err != nil {
		return BuildResult{}, fmt.Errorf("could not create deploy directory: %w", err)
	}

	transaction := watcher.GetTransaction(config.Dir)
	err = build(ctx, projectRoot, config, sch, buildEnv, devMode, transaction)
	if err != nil {
		return BuildResult{}, err
	}
	errors, err := loadProtoErrors(config)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read build errors for module: %w", err)
	}

	result := BuildResult{
		Errors:    errors,
		StartTime: startTime,
	}

	if builderrors.ContainsTerminalError(errors) {
		// skip reading schema
		return result, nil
	}

	moduleSchema, err := schema.ModuleFromProtoFile(config.Schema())
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed to read schema for module: %w", err)
	}
	result.Schema = moduleSchema
	return result, nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) ([]builderrors.Error, error) {
	if _, err := os.Stat(config.Errors); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(config.Errors)
	if err != nil {
		return nil, fmt.Errorf("could not load build errors file: %w", err)
	}

	errorspb := &languagepb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal build errors %w", err)
	}
	return languagepb.ErrorsFromProto(errorspb), nil
}
