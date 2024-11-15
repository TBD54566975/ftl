package languageplugin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/either"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/alecthomas/types/result"
	"google.golang.org/protobuf/types/known/structpb"

	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
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

	// Endpoint of an instance started by the plugin to use in dev mode
	DevEndpoint optional.Option[string]
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
	Plugin *LanguagePlugin
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
	BuildEnv     []string
}

var ErrPluginNotRunning = errors.New("language plugin no longer running")

// PluginFromConfig creates a new language plugin from the given config.
func New(ctx context.Context, dir, language, name string, devMode bool) (p *LanguagePlugin, err error) {
	impl, err := newClientImpl(ctx, dir, language, name)
	if err != nil {
		return nil, err
	}
	return newPluginForTesting(ctx, impl), nil
}

func newPluginForTesting(ctx context.Context, client pluginClient) *LanguagePlugin {
	plugin := &LanguagePlugin{
		client:   client,
		commands: make(chan buildCommand, 64),
		updates:  pubsub.New[PluginEvent](),
	}

	var runCtx context.Context
	runCtx, plugin.cancel = context.WithCancel(ctx)
	go plugin.run(runCtx)
	go plugin.watchForCmdError(runCtx)

	return plugin
}

type buildCommand struct {
	BuildContext
	projectRoot          string
	stubsRoot            string
	rebuildAutomatically bool

	startTime time.Time
	result    chan result.Result[BuildResult]
}

type LanguagePlugin struct {
	client pluginClient

	// cancels the run() context
	cancel context.CancelFunc

	// commands to execute
	commands chan buildCommand

	updates *pubsub.Topic[PluginEvent]
}

// Kill stops the plugin and cleans up any resources.
func (p *LanguagePlugin) Kill() error {
	p.cancel()
	if err := p.client.kill(); err != nil {
		return fmt.Errorf("failed to kill language plugin: %w", err)
	}
	return nil
}

// Updates topic for all update events from the plugin
// The same topic must be returned each time this method is called
func (p *LanguagePlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.updates
}

// GetCreateModuleFlags returns the flags that can be used to create a module for this language.
func (p *LanguagePlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
	res, err := p.client.getCreateModuleFlags(ctx, connect.NewRequest(&langpb.GetCreateModuleFlagsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("failed to get create module flags from plugin: %w", err)
	}
	flags := []*kong.Flag{}
	shorts := map[rune]string{}
	for _, f := range res.Msg.Flags {
		flag := &kong.Flag{
			Value: &kong.Value{
				Name: f.Name,
				Help: f.Help,
				Tag:  &kong.Tag{},
			},
		}
		if f.Envar != nil && *f.Envar != "" {
			flag.Value.Tag.Envs = []string{*f.Envar}
		}
		if f.Default != nil && *f.Default != "" {
			flag.Value.HasDefault = true
			flag.Value.Default = *f.Default
		}
		if f.Short != nil && *f.Short != "" {
			if len(*f.Short) > 1 {
				return nil, fmt.Errorf("invalid flag declared: short flag %q for %v must be a single character", *f.Short, f.Name)
			}
			short := rune((*f.Short)[0])
			if existingFullName, ok := shorts[short]; ok {
				return nil, fmt.Errorf("multiple flags declared with the same short name: %v and %v", existingFullName, f.Name)
			}
			flag.Short = short
			shorts[short] = f.Name

		}
		if f.Placeholder != nil && *f.Placeholder != "" {
			flag.PlaceHolder = *f.Placeholder
		}
		flags = append(flags, flag)
	}
	return flags, nil
}

// CreateModule creates a new module in the given directory with the given name and language.
func (p *LanguagePlugin) CreateModule(ctx context.Context, projConfig projectconfig.Config, moduleConfig moduleconfig.ModuleConfig, flags map[string]string) error {
	genericFlags := map[string]any{}
	for k, v := range flags {
		genericFlags[k] = v
	}
	flagsProto, err := structpb.NewStruct(genericFlags)
	if err != nil {
		return fmt.Errorf("failed to convert flags to proto: %w", err)
	}
	_, err = p.client.createModule(ctx, connect.NewRequest(&langpb.CreateModuleRequest{
		Name:          moduleConfig.Module,
		Dir:           moduleConfig.Dir,
		ProjectConfig: langpb.ProjectConfigToProto(projConfig),
		Flags:         flagsProto,
	}))
	if err != nil {
		return fmt.Errorf("failed to create module: %w", err)
	}
	return nil
}

// ModuleConfigDefaults provides custom defaults for the module config.
//
// The result may be cached by FTL, so defaulting logic should not be changing due to normal module changes.
// For example, it is valid to return defaults based on which build tool is configured within the module directory,
// as that is not expected to change during normal operation.
// It is not recommended to read the module's toml file to determine defaults, as when the toml file is updated,
// the module defaults will not be recalculated.
func (p *LanguagePlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
	resp, err := p.client.moduleConfigDefaults(ctx, connect.NewRequest(&langpb.ModuleConfigDefaultsRequest{
		Dir: dir,
	}))
	if err != nil {
		return moduleconfig.CustomDefaults{}, fmt.Errorf("failed to get module config defaults from plugin: %w", err)
	}
	return customDefaultsFromProto(resp.Msg), nil
}

func customDefaultsFromProto(proto *langpb.ModuleConfigDefaultsResponse) moduleconfig.CustomDefaults {
	return moduleconfig.CustomDefaults{
		DeployDir:          proto.DeployDir,
		Watch:              proto.Watch,
		Build:              optional.Ptr(proto.Build),
		DevModeBuild:       optional.Ptr(proto.DevModeBuild),
		BuildLock:          optional.Ptr(proto.BuildLock),
		GeneratedSchemaDir: optional.Ptr(proto.GeneratedSchemaDir),
		LanguageConfig:     proto.LanguageConfig.AsMap(),
	}
}

// GetDependencies returns the dependencies of the module.
func (p *LanguagePlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, fmt.Errorf("could not convert module config to proto: %w", err)
	}
	resp, err := p.client.getDependencies(ctx, connect.NewRequest(&langpb.DependenciesRequest{
		ModuleConfig: configProto,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies from plugin: %w", err)
	}
	return resp.Msg.Modules, nil
}

// GenerateStubs for the given module.
func (p *LanguagePlugin) GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error {
	moduleProto := module.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	configProto, err := langpb.ModuleConfigToProto(moduleConfig.Abs())
	if err != nil {
		return fmt.Errorf("could not create proto for module config: %w", err)
	}
	var nativeConfigProto *langpb.ModuleConfig
	if config, ok := nativeModuleConfig.Get(); ok {
		nativeConfigProto, err = langpb.ModuleConfigToProto(config.Abs())
		if err != nil {
			return fmt.Errorf("could not create proto for native module config: %w", err)
		}
	}
	_, err = p.client.generateStubs(ctx, connect.NewRequest(&langpb.GenerateStubsRequest{
		Dir:                dir,
		Module:             moduleProto,
		ModuleConfig:       configProto,
		NativeModuleConfig: nativeConfigProto,
	}))
	if err != nil {
		return fmt.Errorf("plugin failed to generate stubs: %w", err)
	}
	return nil
}

// SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
// references to external modules, regardless of whether they are dependencies.
//
// For example, go plugin adds references to all modules into the go.work file so that tools can automatically
// import the modules when users start reference them.
//
// It is optional to do anything with this call.
func (p *LanguagePlugin) SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string) error {
	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return fmt.Errorf("could not create proto for native module config: %w", err)
	}
	_, err = p.client.syncStubReferences(ctx, connect.NewRequest(&langpb.SyncStubReferencesRequest{
		StubsRoot:    dir,
		Modules:      moduleNames,
		ModuleConfig: configProto,
	}))
	if err != nil {
		return fmt.Errorf("plugin failed to sync stub references: %w", err)
	}
	return nil
}

// Build builds the module with the latest config and schema.
// In dev mode, plugin is responsible for automatically rebuilding as relevant files within the module change,
// and publishing these automatic builds updates to Updates().
func (p *LanguagePlugin) Build(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, rebuildAutomatically bool) (BuildResult, error) {
	cmd := buildCommand{
		BuildContext:         bctx,
		projectRoot:          projectRoot,
		stubsRoot:            stubsRoot,
		rebuildAutomatically: rebuildAutomatically,
		startTime:            time.Now(),
		result:               make(chan result.Result[BuildResult]),
	}
	p.commands <- cmd
	select {
	case r := <-cmd.result:
		result, err := r.Result()
		if err != nil {
			return BuildResult{}, err //nolint:wrapcheck
		}
		return result, nil
	case <-ctx.Done():
		return BuildResult{}, fmt.Errorf("error waiting for build to complete: %w", ctx.Err())
	}
}

func (p *LanguagePlugin) watchForCmdError(ctx context.Context) {
	select {
	case err := <-p.client.cmdErr():
		if err == nil {
			// closed
			return
		}
		p.updates.Publish(PluginDiedEvent{
			Plugin: p,
			Error:  err,
		})

	case <-ctx.Done():

	}
}

func (p *LanguagePlugin) run(ctx context.Context) {
	// State
	var bctx BuildContext
	var projectRoot string
	var stubsRoot string

	// if a current build stream is active, this is non-nil
	// this does not indicate if the stream is listening to automatic rebuilds
	var streamChan chan result.Result[*langpb.BuildEvent]
	var streamCancel streamCancelFunc

	// if an explicit build command is active, this is non-nil
	// if this is nil, streamChan may still be open for automatic rebuilds
	var activeBuildCmd optional.Option[buildCommand]

	// if an automatic rebuild was started, this is the time it started
	var autoRebuildStartTime optional.Option[time.Time]

	// build counter is used to generate build request ids
	var contextCounter = 0

	// can not scope logger initially without knowing module name
	logger := log.FromContext(ctx)

	for {
		select {
		// Process incoming commands
		case c := <-p.commands:
			// update state
			contextCounter++
			bctx = c.BuildContext
			projectRoot = c.projectRoot
			stubsRoot = c.stubsRoot

			// module name may have changed, update logger scope
			logger = log.FromContext(ctx).Scope(bctx.Config.Module)

			if _, ok := activeBuildCmd.Get(); ok {
				c.result <- result.Err[BuildResult](fmt.Errorf("build already in progress"))
				continue
			}
			configProto, err := langpb.ModuleConfigToProto(bctx.Config.Abs())
			if err != nil {
				c.result <- result.Err[BuildResult](err)
				continue
			}

			schemaProto := bctx.Schema.ToProto().(*schemapb.Schema) //nolint:forcetypeassert

			if streamChan != nil {
				// tell plugin about new build context so that it rebuilds in existing build stream
				_, err = p.client.buildContextUpdated(ctx, connect.NewRequest(&langpb.BuildContextUpdatedRequest{
					BuildContext: &langpb.BuildContext{
						Id:           contextID(bctx.Config, contextCounter),
						ModuleConfig: configProto,
						Schema:       schemaProto,
						Dependencies: bctx.Dependencies,
						BuildEnv:     c.BuildEnv,
					},
				}))
				if err != nil {
					c.result <- result.Err[BuildResult](fmt.Errorf("failed to send updated build context to plugin: %w", err))
					continue
				}
				activeBuildCmd = optional.Some[buildCommand](c)
				continue
			}

			newStreamChan, newCancelFunc, err := p.client.build(ctx, connect.NewRequest(&langpb.BuildRequest{
				ProjectRoot:          projectRoot,
				StubsRoot:            stubsRoot,
				RebuildAutomatically: c.rebuildAutomatically,
				BuildContext: &langpb.BuildContext{
					Id:           contextID(bctx.Config, contextCounter),
					ModuleConfig: configProto,
					Schema:       schemaProto,
					Dependencies: bctx.Dependencies,
					BuildEnv:     c.BuildEnv,
				},
			}))
			if err != nil {
				c.result <- result.Err[BuildResult](fmt.Errorf("failed to start build stream: %w", err))
				continue
			}
			activeBuildCmd = optional.Some[buildCommand](c)
			streamChan = newStreamChan
			streamCancel = newCancelFunc

		// Receive messages from the current build stream
		case r := <-streamChan:
			e, err := r.Result()
			if err != nil {
				// Stream failed
				if c, ok := activeBuildCmd.Get(); ok {
					c.result <- result.Err[BuildResult](err)
					activeBuildCmd = optional.None[buildCommand]()
				}
				streamCancel = nil
				streamChan = nil
			}
			if e == nil {
				streamChan = nil
				streamCancel = nil
				continue
			}

			switch e.Event.(type) {
			case *langpb.BuildEvent_AutoRebuildStarted:
				if _, ok := activeBuildCmd.Get(); ok {
					logger.Debugf("ignoring automatic rebuild started during explicit build")
					continue
				}
				autoRebuildStartTime = optional.Some(time.Now())
				p.updates.Publish(AutoRebuildStartedEvent{
					Module: bctx.Config.Module,
				})
			case *langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure:
				streamEnded := false
				cmdEnded := false
				result, eventContextID, isAutomaticRebuild := getBuildSuccessOrFailure(e)
				if activeBuildCmd.Ok() == isAutomaticRebuild {
					if isAutomaticRebuild {
						logger.Debugf("ignoring automatic rebuild while expecting explicit build")
					} else {
						// This is likely a language plugin bug, but we can ignore it
						logger.Warnf("ignoring explicit build while none was requested")
					}
					continue
				} else if eventContextID != contextID(bctx.Config, contextCounter) {
					logger.Debugf("received build for outdated context %q; expected %q", eventContextID, contextID(bctx.Config, contextCounter))
					continue
				}

				var startTime time.Time
				if cmd, ok := activeBuildCmd.Get(); ok {
					startTime = cmd.startTime
				} else if t, ok := autoRebuildStartTime.Get(); ok {
					startTime = t
				} else {
					// Plugin did not declare when it started to build.
					startTime = time.Now()
				}
				streamEnded, cmdEnded = p.handleBuildResult(bctx.Config.Module, result, activeBuildCmd, startTime)
				if streamEnded {
					streamCancel()
					streamCancel = nil
					streamChan = nil
				}
				if cmdEnded {
					activeBuildCmd = optional.None[buildCommand]()
				}
			}

		case <-ctx.Done():
			if streamCancel != nil {
				streamCancel()
			}
			return
		}
	}
}

// getBuildSuccessOrFailure takes a BuildFailure or BuildSuccess event and returns the shared fields and an either wrapped result.
// This makes it easier to have some shared logic for both event types.
func getBuildSuccessOrFailure(e *langpb.BuildEvent) (result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure], contextID string, isAutomaticRebuild bool) {
	switch e := e.Event.(type) {
	case *langpb.BuildEvent_BuildSuccess:
		return either.LeftOf[*langpb.BuildEvent_BuildFailure](e), e.BuildSuccess.ContextId, e.BuildSuccess.IsAutomaticRebuild
	case *langpb.BuildEvent_BuildFailure:
		return either.RightOf[*langpb.BuildEvent_BuildSuccess](e), e.BuildFailure.ContextId, e.BuildFailure.IsAutomaticRebuild
	default:
		panic(fmt.Sprintf("unexpected event type %T", e))
	}
}

// handleBuildResult processes the result of a build and publishes the appropriate events.
func (p *LanguagePlugin) handleBuildResult(module string, r either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure],
	activeBuildCmd optional.Option[buildCommand], startTime time.Time) (streamEnded, cmdEnded bool) {
	buildResult, err := buildResultFromProto(r, startTime)
	if cmd, ok := activeBuildCmd.Get(); ok {
		// handle explicit build
		cmd.result <- result.From(buildResult, err)

		cmdEnded = true
		if !cmd.rebuildAutomatically {
			streamEnded = true
		}
		return
	}
	// handle auto rebuild
	p.updates.Publish(AutoRebuildEndedEvent{
		Module: module,
		Result: result.From(buildResult, err),
	})
	return
}

func buildResultFromProto(result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure], startTime time.Time) (buildResult BuildResult, err error) {
	switch result := result.(type) {
	case either.Left[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]:
		buildSuccess := result.Get().BuildSuccess

		moduleSch, err := schema.ModuleFromProto(buildSuccess.Module)
		if err != nil {
			return BuildResult{}, fmt.Errorf("failed to parse schema: %w", err)
		}
		if moduleSch.Runtime != nil && len(strings.Split(moduleSch.Runtime.Image, ":")) != 1 {
			return BuildResult{}, fmt.Errorf("image tag not supported in runtime image: %s", moduleSch.Runtime.Image)
		}

		errs := langpb.ErrorsFromProto(buildSuccess.Errors)
		builderrors.SortErrorsByPosition(errs)
		return BuildResult{
			Errors:      errs,
			Schema:      moduleSch,
			Deploy:      buildSuccess.Deploy,
			StartTime:   startTime,
			DevEndpoint: optional.Ptr(buildSuccess.DevEndpoint),
		}, nil
	case either.Right[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]:
		buildFailure := result.Get().BuildFailure

		errs := langpb.ErrorsFromProto(buildFailure.Errors)
		builderrors.SortErrorsByPosition(errs)

		if !builderrors.ContainsTerminalError(errs) {
			// This happens if the language plugin returns BuildFailure but does not include any errors with level ERROR.
			// Language plugins should always include at least one error with level ERROR in the case of a build failure.
			errs = append(errs, builderrors.Error{
				Msg:   "unexpected build failure without error level ERROR",
				Level: builderrors.ERROR,
				Type:  builderrors.FTL,
			})
		}

		return BuildResult{
			StartTime:              startTime,
			Errors:                 errs,
			InvalidateDependencies: buildFailure.InvalidateDependencies,
		}, nil
	default:
		panic(fmt.Sprintf("unexpected result type %T", result))
	}
}

func contextID(config moduleconfig.ModuleConfig, counter int) string {
	return fmt.Sprintf("%v-%v", config.Module, counter)
}
