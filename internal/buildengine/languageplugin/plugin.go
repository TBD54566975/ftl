package languageplugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/alecthomas/types/result"

	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute

type BuildResult struct {
	StartTime time.Time

	Schema *schema.Module
	Errors []builderrors.Error

	// Files to deploy, relative to the module config's DeployDir
	Deploy []string

	// Whether the module needs to recalculate its dependencies
	InvalidateDependencies bool
}

// PluginEvent is used to notify of updates from the plugin.
//
//sumtype:decl
type PluginEvent interface {
	pluginEvent()
}

type PluginBuildEvent interface {
	PluginEvent
	ModuleName() string
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
	Result result.Result[BuildResult]
}

func (AutoRebuildEndedEvent) pluginEvent()         {}
func (e AutoRebuildEndedEvent) ModuleName() string { return e.Module }

// PluginDiedEvent is sent when the plugin dies.
type PluginDiedEvent struct {
	// Plugins do not always have an associated module name, so we include the module
	Plugin LanguagePlugin
	Error  error
}

func (PluginDiedEvent) pluginEvent() {}

// BuildContext contains contextual information needed to build.
//
// Any change to the build context would require a new build.
type BuildContext struct {
	Config       moduleconfig.ModuleConfig
	Schema       *schema.Schema
	Dependencies []string
}

var ErrPluginNotRunning = errors.New("language plugin no longer running")

// LanguagePlugin handles building and scaffolding modules in a specific language.
type LanguagePlugin interface {
	// Updates topic for all update events from the plugin
	// The same topic must be returned each time this method is called
	Updates() *pubsub.Topic[PluginEvent]

	// GetModuleConfigDefaults provides custom defaults for the module config.
	//
	// The result may be cached by FTL, so defaulting logic should not be changing due to normal module changes.
	// For example, it is valid to return defaults based on which build tool is configured within the module directory,
	// as that is not expected to change during normal operation.
	// It is not recommended to read the module's toml file to determine defaults, as when the toml file is updated,
	// the module defaults will not be recalculated.
	ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error)

	// GetCreateModuleFlags returns the flags that can be used to create a module for this language.
	GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error)

	// CreateModule creates a new module in the given directory with the given name and language.
	CreateModule(ctx context.Context, projConfig projectconfig.Config, moduleConfig moduleconfig.ModuleConfig, flags map[string]string) error

	// GetDependencies returns the dependencies of the module.
	GetDependencies(ctx context.Context, moduleConfig moduleconfig.ModuleConfig) ([]string, error)

	// Build builds the module with the latest config and schema.
	// In dev mode, plugin is responsible for automatically rebuilding as relevant files within the module change,
	// and publishing these automatic builds updates to Updates().
	Build(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, rebuildAutomatically bool) (BuildResult, error)

	// Generate stubs for the given module.
	GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error

	// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
	// references to external modules, regardless of whether they are dependencies.
	//
	// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
	// import the modules when users start reference them.
	//
	// It is optional to do anything with this call.
	SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string) error

	// Kill stops the plugin and cleans up any resources.
	Kill() error
}

// PluginFromConfig creates a new language plugin from the given config.
func New(ctx context.Context, bindAllocator *bind.BindAllocator, language string) (p LanguagePlugin, err error) {
	switch language {
	case "go":
		return newGoPlugin(ctx), nil
	case "java", "kotlin":
		return newJavaPlugin(ctx, language), nil
	case "rust":
		return newRustPlugin(ctx), nil
	default:
		return newExternalPlugin(ctx, bindAllocator.Next(), language)
	}
}

//sumtype:decl
type pluginCommand interface {
	pluginCmd()
}

type buildCommand struct {
	BuildContext
	projectRoot          string
	stubsRoot            string
	buildEnv             []string
	rebuildAutomatically bool

	result chan either.Either[BuildResult, error]
}

func (buildCommand) pluginCmd() {}

type dependenciesFunc func() ([]string, error)
type getDependenciesCommand struct {
	dependenciesFunc dependenciesFunc

	result chan either.Either[[]string, error]
}

func (getDependenciesCommand) pluginCmd() {}

type buildFunc = func(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, rebuildAutomatically bool, transaction watch.ModifyFilesTransaction) (BuildResult, error)

type CompilerBuildError struct {
	err error
}

func (e CompilerBuildError) Error() string {
	return e.err.Error()
}

func (e CompilerBuildError) Unwrap() error {
	return e.err
}

// internalPlugin is used by languages that have not been split off into their own external plugins yet.
// It has standard behaviours around building and watching files.
type internalPlugin struct {
	language string

	// build is called when a new build is explicitly requested or when a watched file changes
	buildFunc buildFunc

	// commands to execute
	commands chan pluginCommand

	updates *pubsub.Topic[PluginEvent]
	cancel  context.CancelFunc
}

func newInternalPlugin(ctx context.Context, language string, build buildFunc) *internalPlugin {
	plugin := &internalPlugin{
		language:  language,
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

func (p *internalPlugin) Kill() error {
	p.cancel()
	return nil
}

func (p *internalPlugin) GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error {
	return nil
}

func (p *internalPlugin) SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string) error {
	return nil
}

func (p *internalPlugin) Build(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, rebuildAutomatically bool) (BuildResult, error) {
	cmd := buildCommand{
		BuildContext:         bctx,
		projectRoot:          projectRoot,
		stubsRoot:            stubsRoot,
		buildEnv:             buildEnv,
		rebuildAutomatically: rebuildAutomatically,
		result:               make(chan either.Either[BuildResult, error]),
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
	var watcher *watch.Watcher
	watchChan := make(chan watch.WatchEvent, 128)

	// State
	// This is updated when given explicit build commands and used for automatic rebuilds
	var bctx BuildContext
	var projectRoot string
	var stubsRoot string
	var buildEnv []string
	watching := false

	for {
		select {
		case cmd := <-p.commands:
			switch c := cmd.(type) {
			case buildCommand:
				// update state
				bctx = c.BuildContext
				projectRoot = c.projectRoot
				stubsRoot = c.stubsRoot
				buildEnv = c.buildEnv

				if watcher == nil {
					watcher = watch.NewWatcher(bctx.Config.Watch...)
				}

				// begin watching if needed
				if c.rebuildAutomatically && !watching {
					watching = true
					topic, err := watcher.Watch(ctx, time.Second, []string{bctx.Config.Abs().Dir})
					if err != nil {
						c.result <- either.RightOf[BuildResult](fmt.Errorf("failed to start watching: %w", err))
						continue
					}
					topic.Subscribe(watchChan)
				}

				// build
				result, err := buildAndLoadResult(ctx, projectRoot, stubsRoot, bctx, buildEnv, c.rebuildAutomatically, watcher, p.buildFunc)
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
			case watch.WatchEventModuleChanged:
				// automatic rebuild

				p.updates.Publish(AutoRebuildStartedEvent{Module: bctx.Config.Module})
				p.updates.Publish(AutoRebuildEndedEvent{
					Module: bctx.Config.Module,
					Result: result.From(buildAndLoadResult(ctx, projectRoot, stubsRoot, bctx, buildEnv, true, watcher, p.buildFunc)),
				})
			case watch.WatchEventModuleAdded:
				// ignore

			case watch.WatchEventModuleRemoved:
				// ignore
			}

		case <-ctx.Done():
			return
		}
	}
}

func buildAndLoadResult(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, devMode bool, watcher *watch.Watcher, build buildFunc) (BuildResult, error) {
	config := bctx.Config.Abs()
	release, err := flock.Acquire(ctx, config.BuildLock, BuildLockTimeout)
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
	result, err := build(ctx, projectRoot, stubsRoot, bctx, buildEnv, devMode, transaction)
	if err != nil {
		return BuildResult{}, err
	}
	result.StartTime = startTime
	return result, nil
}
