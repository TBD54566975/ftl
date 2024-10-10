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

//sumtype:decl
type externalPluginCommand interface {
	externalPluginCmd()
}

type externalBuildCommand struct {
	BuildContext
	projectRoot          string
	rebuildAutomatically bool

	result chan either.Either[BuildResult, error]
}

func (externalBuildCommand) externalPluginCmd() {}

type externalPlugin struct {
	client externalPluginClient

	// cancels the run() context
	cancel context.CancelFunc

	// commands to execute
	commands chan externalPluginCommand

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
		commands: make(chan externalPluginCommand, 64),
		updates:  pubsub.New[PluginEvent](),
	}

	var runCtx context.Context
	runCtx, plugin.cancel = context.WithCancel(ctx)
	go plugin.run(runCtx)

	return plugin
}

func (p *externalPlugin) Kill() error {
	p.cancel()
	return p.client.kill()
}

func (p *externalPlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.updates
}

func (p *externalPlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
	res, err := p.client.getCreateModuleFlags(ctx, connect.NewRequest(&langpb.GetCreateModuleFlagsRequest{}))
	if err != nil {
		return nil, err
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
	_, err := p.client.createModule(ctx, connect.NewRequest(&langpb.CreateModuleRequest{
		Name: moduleConfig.Module,
		Path: moduleConfig.Dir,
		ProjectConfig: &langpb.ProjectConfig{
			NoGit:  projConfig.NoGit,
			Hermit: projConfig.Hermit,
		},
	}))
	return err
}

func (p *externalPlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
	resp, err := p.client.moduleConfigDefaults(ctx, connect.NewRequest(&langpb.ModuleConfigDefaultsRequest{
		Path: dir,
	}))
	if err != nil {
		return moduleconfig.CustomDefaults{}, err
	}
	return moduleconfig.CustomDefaults{
		DeployDir:          resp.Msg.DeployDir,
		Watch:              resp.Msg.Watch,
		Build:              optional.Ptr(resp.Msg.Build),
		GeneratedSchemaDir: optional.Ptr(resp.Msg.GeneratedSchemaDir),
		LanguageConfig:     resp.Msg.LanguageConfig.AsMap(),
	}, nil
}

func (p *externalPlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	configProto, err := protoFromModuleConfig(config)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.getDependencies(ctx, connect.NewRequest(&langpb.DependenciesRequest{
		ModuleConfig: configProto,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Modules, nil
}

// Build may result in a Build or BuildContextUpdated grpc call with the plugin, depending if a build stream is already set up
func (p *externalPlugin) Build(ctx context.Context, projectRoot string, bctx BuildContext, buildEnv []string, rebuildAutomatically bool) (BuildResult, error) {
	cmd := externalBuildCommand{
		BuildContext:         bctx,
		projectRoot:          projectRoot,
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

func protoFromModuleConfig(c moduleconfig.ModuleConfig) (*langpb.ModuleConfig, error) {
	config := c.Abs()
	proto := &langpb.ModuleConfig{
		Name:      config.Module,
		Path:      config.Dir,
		DeployDir: config.DeployDir,
		Watch:     config.Watch,
	}
	if config.Build != "" {
		proto.Build = &config.Build
	}
	if config.GeneratedSchemaDir != "" {
		proto.GeneratedSchemaDir = &config.GeneratedSchemaDir
	}

	langConfigProto, err := structpb.NewStruct(config.LanguageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal language config: %w", err)
	}
	proto.LanguageConfig = langConfigProto

	return proto, nil
}

func (p *externalPlugin) run(ctx context.Context) {
	// State
	var bctx BuildContext
	var projectRoot string

	// if a current build stream is active, this is non-nil
	// this does not indicate if the stream is listening to automatic rebuilds
	var streamChan chan *langpb.BuildEvent
	var streamCancel streamCancelFunc

	// if an explicit build command is active, this is non-nil
	// if this is nil, streamChan may still be open for automatic rebuilds
	var activeBuildCmd optional.Option[externalBuildCommand]

	// build counter is used to generate build request ids
	var contextCounter = 0

	logger := log.FromContext(ctx)

	for {
		select {
		// Process incoming commands
		case cmd := <-p.commands:
			switch c := cmd.(type) {
			case externalBuildCommand:
				// update state
				projectRoot = c.projectRoot
				bctx = c.BuildContext
				logger = log.FromContext(ctx).Scope(bctx.Config.Module)
				if _, ok := activeBuildCmd.Get(); ok {
					c.result <- either.RightOf[BuildResult](fmt.Errorf("build already in progress"))
					continue
				}
				configProto, err := protoFromModuleConfig(bctx.Config)
				if err != nil {
					c.result <- either.RightOf[BuildResult](err)
					continue
				}

				activeBuildCmd = optional.Some[externalBuildCommand](c)
				contextCounter++

				if streamChan != nil {
					// tell plugin about new build context so that it rebuilds in existing build stream
					p.client.buildContextUpdated(ctx, connect.NewRequest(&langpb.BuildContextUpdatedRequest{
						BuildContext: &langpb.BuildContext{
							Id:           contextId(bctx.Config, contextCounter),
							ModuleConfig: configProto,
							Schema:       bctx.Schema.ToProto().(*schemapb.Schema), //nolint:forcetypeassert
							Dependencies: bctx.Dependencies,
						},
					}))
					continue
				}

				newStreamChan, newCancelFunc, err := p.client.build(ctx, connect.NewRequest(&langpb.BuildRequest{
					ProjectPath:          projectRoot,
					RebuildAutomatically: c.rebuildAutomatically,
					BuildContext: &langpb.BuildContext{
						Id:           contextId(bctx.Config, contextCounter),
						ModuleConfig: configProto,
						Schema:       bctx.Schema.ToProto().(*schemapb.Schema), //nolint:forcetypeassert
						Dependencies: bctx.Dependencies,
					},
				}))
				if err != nil {
					// TODO: error
					continue
				}
				streamChan = newStreamChan
				streamCancel = newCancelFunc
			}

		// Receive messages from the current build stream
		case e := <-streamChan:
			if e == nil {
				streamChan = nil
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
				result, eventContextId, isAutomaticRebuild := getBuildSuccessOrFailure(e)
				if activeBuildCmd.Ok() == isAutomaticRebuild {
					logger.Debugf("ignoring automatic rebuild while expecting explicit build")
					continue
				} else if eventContextId != contextId(bctx.Config, contextCounter) {
					logger.Debugf("received build for outdated context %q; expected %q", eventContextId, contextId(bctx.Config, contextCounter))
					continue
				}
				streamEnded, cmdEnded = p.handleBuildResult(bctx.Config.Module, result, activeBuildCmd)
				if streamEnded {
					streamCancel()
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
func getBuildSuccessOrFailure(e *langpb.BuildEvent) (result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure], contextId string, isAutomaticRebuild bool) {
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
func (p *externalPlugin) handleBuildResult(module string, result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure], activeBuildCmd optional.Option[externalBuildCommand]) (streamEnded, cmdEnded bool) {
	buildResult, err := buildResultFromProto(result)
	if cmd, ok := activeBuildCmd.Get(); ok {
		// handle explicit build
		if err != nil {
			cmd.result <- either.RightOf[BuildResult](err)
		} else {
			cmd.result <- either.LeftOf[error](buildResult)
		}
		cmdEnded = true
		if !cmd.rebuildAutomatically {
			streamEnded = true
		}
		return
	}
	// handle auto rebuild
	if err != nil {
		p.updates.Publish(AutoRebuildEndedEvent{
			Module: module,
			Result: either.RightOf[BuildResult](err),
		})
	} else {
		p.updates.Publish(AutoRebuildEndedEvent{
			Module: module,
			Result: either.LeftOf[error](buildResult),
		})
	}
	return
}

func buildResultFromProto(result either.Either[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]) (buildResult BuildResult, err error) {
	switch result := result.(type) {
	case either.Left[*langpb.BuildEvent_BuildSuccess, *langpb.BuildEvent_BuildFailure]:
		buildSuccess := result.Get().BuildSuccess
		var moduleSch *schema.Module
		if buildSuccess.Module != nil {
			sch, err := schema.ModuleFromProto(buildSuccess.Module)
			if err != nil {
				return BuildResult{}, fmt.Errorf("failed to parse schema: %w", err)
			}
			moduleSch = sch
		}

		errs := langpb.ErrorsFromProto(buildSuccess.Errors)
		builderrors.SortErrorsByPosition(errs)
		return BuildResult{
			Errors: errs,
			Schema: moduleSch,
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

func contextId(config moduleconfig.ModuleConfig, counter int) string {
	return fmt.Sprintf("%v-%v", config.Module, counter)
}
