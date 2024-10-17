package languageplugin

import (
	"context"
	"fmt"
	"net/url"
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

const launchTimeout = 10 * time.Second

type externalBuildCommand struct {
	BuildContext
	projectRoot          string
	stubsRoot            string
	rebuildAutomatically bool

	startTime time.Time
	result    chan result.Result[BuildResult]
}

type externalPlugin struct {
	client externalPluginClient

	// cancels the run() context
	cancel context.CancelFunc

	// commands to execute
	commands chan externalBuildCommand

	updates *pubsub.Topic[PluginEvent]
}

var _ LanguagePlugin = &externalPlugin{}

func newExternalPlugin(ctx context.Context, bind *url.URL, language string) (*externalPlugin, error) {
	impl, err := newExternalPluginImpl(ctx, bind, language)
	if err != nil {
		return nil, err
	}
	return newExternalPluginForTesting(ctx, impl), nil
}

func newExternalPluginForTesting(ctx context.Context, client externalPluginClient) *externalPlugin {
	plugin := &externalPlugin{
		client:   client,
		commands: make(chan externalBuildCommand, 64),
		updates:  pubsub.New[PluginEvent](),
	}

	var runCtx context.Context
	runCtx, plugin.cancel = context.WithCancel(ctx)
	go plugin.run(runCtx)

	return plugin
}

func (p *externalPlugin) Kill() error {
	p.cancel()
	if err := p.client.kill(); err != nil {
		return fmt.Errorf("failed to kill language plugin: %w", err)
	}
	return nil
}

func (p *externalPlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.updates
}

func (p *externalPlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
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
func (p *externalPlugin) CreateModule(ctx context.Context, projConfig projectconfig.Config, moduleConfig moduleconfig.ModuleConfig, flags map[string]string) error {
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

func (p *externalPlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
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
		GeneratedSchemaDir: optional.Ptr(proto.GeneratedSchemaDir),
		LanguageConfig:     proto.LanguageConfig.AsMap(),
	}
}

func (p *externalPlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	configProto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, err
	}
	resp, err := p.client.getDependencies(ctx, connect.NewRequest(&langpb.DependenciesRequest{
		ModuleConfig: configProto,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies from plugin: %w", err)
	}
	return resp.Msg.Modules, nil
}

func (p *externalPlugin) GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error {
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

func (p *externalPlugin) SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string) error {
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

// Build may result in a Build or BuildContextUpdated grpc call with the plugin, depending if a build stream is already set up
func (p *externalPlugin) Build(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, rebuildAutomatically bool) (BuildResult, error) {
	cmd := externalBuildCommand{
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
		result.StartTime = cmd.startTime
		return result, nil
	case <-ctx.Done():
		return BuildResult{}, fmt.Errorf("error waiting for build to complete: %w", ctx.Err())
	}
}

func (p *externalPlugin) run(ctx context.Context) {
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
	var activeBuildCmd optional.Option[externalBuildCommand]

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
					},
				}))
				if err != nil {
					c.result <- result.Err[BuildResult](fmt.Errorf("failed to send updated build context to plugin: %w", err))
					continue
				}
				activeBuildCmd = optional.Some[externalBuildCommand](c)
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
				},
			}))
			if err != nil {
				c.result <- result.Err[BuildResult](fmt.Errorf("failed to start build stream: %w", err))
				continue
			}
			activeBuildCmd = optional.Some[externalBuildCommand](c)
			streamChan = newStreamChan
			streamCancel = newCancelFunc

		// Receive messages from the current build stream
		case r := <-streamChan:
			e, err := r.Result()
			if err != nil {
				// Stream failed
				if c, ok := activeBuildCmd.Get(); ok {
					c.result <- result.Err[BuildResult](err)
					activeBuildCmd = optional.None[externalBuildCommand]()
				}
				streamCancel = nil
				streamChan = nil
			}
			if e == nil {
				streamChan = nil
				streamCancel = nil
				continue
			}

			switch event := e.Event.(type) {
			case *langpb.BuildEvent_LogMessage:
				logger.Logf(langpb.LogLevelFromProto(event.LogMessage.Level), "%s", event.LogMessage.Message)
			case *langpb.BuildEvent_AutoRebuildStarted:
				if _, ok := activeBuildCmd.Get(); ok {
					logger.Debugf("ignoring automatic rebuild started during explicit build")
					continue
				}
				p.updates.Publish(AutoRebuildStartedEvent{
					Module: bctx.Config.Module,
				})
			case *langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure:
				streamEnded := false
				cmdEnded := false
				result, eventContextID, isAutomaticRebuild := getBuildSuccessOrFailure(e)
				if activeBuildCmd.Ok() == isAutomaticRebuild {
					logger.Debugf("ignoring automatic rebuild while expecting explicit build")
					continue
				} else if eventContextID != contextID(bctx.Config, contextCounter) {
					logger.Debugf("received build for outdated context %q; expected %q", eventContextID, contextID(bctx.Config, contextCounter))
					continue
				}
				streamEnded, cmdEnded = p.handleBuildResult(bctx.Config.Module, result, activeBuildCmd)
				if streamEnded {
					streamCancel()
					streamCancel = nil
					streamChan = nil
				}
				if cmdEnded {
					activeBuildCmd = optional.None[externalBuildCommand]()
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
func (p *externalPlugin) handleBuildResult(module string, r either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure], activeBuildCmd optional.Option[externalBuildCommand]) (streamEnded, cmdEnded bool) {
	buildResult, err := buildResultFromProto(r)
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

func buildResultFromProto(result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]) (buildResult BuildResult, err error) {
	switch result := result.(type) {
	case either.Left[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]:
		buildSuccess := result.Get().BuildSuccess

		moduleSch, err := schema.ModuleFromProto(buildSuccess.Module)
		if err != nil {
			return BuildResult{}, fmt.Errorf("failed to parse schema: %w", err)
		}

		errs := langpb.ErrorsFromProto(buildSuccess.Errors)
		builderrors.SortErrorsByPosition(errs)
		return BuildResult{
			Errors: errs,
			Schema: moduleSch,
			Deploy: buildSuccess.Deploy,
		}, nil
	case either.Right[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]:
		buildFailure := result.Get().BuildFailure

		errs := langpb.ErrorsFromProto(buildFailure.Errors)
		builderrors.SortErrorsByPosition(errs)
		return BuildResult{
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
