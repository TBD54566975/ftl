package goplugin

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	langconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	islices "github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/watch"
)

const BuildLockTimeout = time.Minute

//sumtype:decl
type updateEvent interface{ updateEvent() }

type buildContextUpdatedEvent struct {
	buildCtx buildContext
}

func (buildContextUpdatedEvent) updateEvent() {}

type filesUpdatedEvent struct{}

func (filesUpdatedEvent) updateEvent() {}

// buildContext contains contextual information needed to build.
type buildContext struct {
	Id           string
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, err
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	return buildContext{
		Id:           proto.Id,
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
	}, nil
}

type Service struct {
	updatesTopic          *pubsub.Topic[updateEvent]
	acceptsContextUpdates atomic.Value[bool]
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New() *Service {
	return &Service{
		updatesTopic: pubsub.New[updateEvent](),
	}
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	log.FromContext(ctx).Infof("Received Ping")
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) GetCreateModuleFlags(ctx context.Context, req *connect.Request[langpb.GetCreateModuleFlagsRequest]) (*connect.Response[langpb.GetCreateModuleFlagsResponse], error) {
	return connect.NewResponse(&langpb.GetCreateModuleFlagsResponse{
		Flags: []*langpb.GetCreateModuleFlagsResponse_Flag{
			{
				Name:        "replace",
				Help:        "Replace a module import path with a local path in the initialised FTL module.",
				Envar:       optional.Some("FTL_INIT_GO_REPLACE").Ptr(),
				Short:       optional.Some("r").Ptr(),
				Placeholder: optional.Some("OLD=NEW,...").Ptr(),
			},
		},
	}), nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

// Generates files for a new module with the requested name
func (s *Service) CreateModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	logger := log.FromContext(ctx)
	flags := req.Msg.Flags.AsMap()
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)

	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := scaffoldingContext{
		Name:      req.Msg.Name,
		GoVersion: runtime.Version()[2:],
		Replace:   map[string]string{},
	}
	if replaceValue, ok := flags["replace"]; ok && replaceValue != "" {
		replaceStr, ok := replaceValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid replace flag is not a string: %v", replaceValue)
		}
		for _, replace := range strings.Split(replaceStr, ",") {
			parts := strings.Split(replace, "=")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid replace flag (format: A=B,C=D): %q", replace)
			}
			sctx.Replace[parts[0]] = parts[1]
		}
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy: %s", req.Msg.Dir)
	if err := exec.Command(ctx, log.Debug, req.Msg.Dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return nil, fmt.Errorf("could not tidy: %w", err)
	}
	return connect.NewResponse(&langpb.CreateModuleResponse{}), nil
}

// Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
func (s *Service) ModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	deployDir := ".ftl"
	watch := []string{"**/*.go", "go.mod", "go.sum"}
	additionalWatch, err := replacementWatches(req.Msg.Dir, deployDir)
	watch = append(watch, additionalWatch...)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&langpb.ModuleConfigDefaultsResponse{
		Watch:     watch,
		DeployDir: deployDir,
	}), nil
}

// Extract dependencies for a module
// FTL will ensure that these dependencies are built before requesting a build for this module.
func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	deps, err := compile.ExtractDependencies(config)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&langpb.DependenciesResponse{
		Modules: deps,
	}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	moduleSchema, err := schema.ModuleFromProto(req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("could not parse schema: %w", err)
	}
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	var nativeConfig optional.Option[moduleconfig.AbsModuleConfig]
	if req.Msg.NativeModuleConfig != nil {
		nativeConfig = optional.Some(langpb.ModuleConfigFromProto(req.Msg.NativeModuleConfig))
	}

	err = compile.GenerateStubs(ctx, req.Msg.Dir, moduleSchema, config, nativeConfig)
	if err != nil {
		return nil, fmt.Errorf("could not generate stubs: %w", err)
	}
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
	config := langpb.ModuleConfigFromProto(req.Msg.ModuleConfig)
	err := compile.SyncGeneratedStubReferences(ctx, config, req.Msg.StubsRoot, req.Msg.Modules)
	if err != nil {
		return nil, fmt.Errorf("could not sync stub references: %w", err)
	}
	return connect.NewResponse(&langpb.SyncStubReferencesResponse{}), nil
}

// Build the module and stream back build events.
//
// A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
// end of the build.
//
// The request can include the option to "rebuild_automatically". In this case the plugin should watch for
// file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
// rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
// calls.
func (s *Service) Build(ctx context.Context, req *connect.Request[langpb.BuildRequest], stream *connect.ServerStream[langpb.BuildEvent]) error {
	events := make(chan updateEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)

	// TODO: can i create a logger than streams events?
	// TODO: do we need a cancellable context?
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// stream logs to FTL
	ctx = log.ContextWithLogger(ctx, newLoggerForStream(log.Debug, stream))

	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return err
	}

	watchPatterns, err := relativeWatchPatterns(buildCtx.Config.Dir, buildCtx.Config.Watch)
	if err != nil {
		return err
	}

	watcher := watch.NewWatcher(watchPatterns...)
	if req.Msg.RebuildAutomatically {
		s.acceptsContextUpdates.Store(true)
		defer s.acceptsContextUpdates.Store(false)

		if err := watchFiles(ctx, watcher, buildCtx, events); err != nil {
			return err
		}
	}

	// Initial build
	if err := buildAndSend(ctx, stream, req.Msg.ProjectRoot, req.Msg.StubsRoot, buildCtx, false, watcher.GetTransaction(buildCtx.Config.Dir)); err != nil {
		return err
	}
	if !req.Msg.RebuildAutomatically {
		return nil
	}

	// Watch for changes and build as needed
	for {
		select {
		case e := <-events:
			var isAutomaticRebuild bool
			buildCtx, isAutomaticRebuild = buildContextFromPendingEvents(ctx, buildCtx, events, e)
			log.FromContext(ctx).Infof("Building (auto = %v) with context id: %s", isAutomaticRebuild, buildCtx.Id)
			if isAutomaticRebuild {
				err = stream.Send(&langpb.BuildEvent{
					Event: &langpb.BuildEvent_AutoRebuildStarted{
						AutoRebuildStarted: &langpb.AutoRebuildStarted{
							ContextId: buildCtx.Id,
						},
					},
				})
				if err != nil {
					return fmt.Errorf("could not send auto rebuild started event: %w", err)
				}
			}
			if err = buildAndSend(ctx, stream, req.Msg.ProjectRoot, req.Msg.StubsRoot, buildCtx, isAutomaticRebuild, watcher.GetTransaction(buildCtx.Config.Dir)); err != nil {
				return err
			}
		case <-ctx.Done():
			log.FromContext(ctx).Infof("Build call ending - ctx cancelled")
			return nil
		}
	}
}

// While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
// build context is updated.
//
// Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
// event with the updated build context id with "is_automatic_rebuild" as false.
func (s *Service) BuildContextUpdated(ctx context.Context, req *connect.Request[langpb.BuildContextUpdatedRequest]) (*connect.Response[langpb.BuildContextUpdatedResponse], error) {
	if !s.acceptsContextUpdates.Load() {
		return nil, fmt.Errorf("plugin does not accept context updates because these is no build stream allowing rebuilds")
	}
	buildCtx, err := buildContextFromProto(req.Msg.BuildContext)
	if err != nil {
		return nil, err
	}

	s.updatesTopic.Publish(buildContextUpdatedEvent{
		buildCtx: buildCtx,
	})

	return connect.NewResponse(&langpb.BuildContextUpdatedResponse{}), nil
}

func watchFiles(ctx context.Context, watcher *watch.Watcher, buildCtx buildContext, events chan updateEvent) error {
	// TODO: add back patterns from module config via grpc
	// TODO: do we need to stop the watcher when the build call ends? or is that automatic?
	watchTopic, err := watcher.Watch(ctx, time.Second, []string{buildCtx.Config.Dir})
	if err != nil {
		return fmt.Errorf("could not watch for file changes: %w", err)
	}
	log.FromContext(ctx).Infof("Watching for file changes: %s", buildCtx.Config.Dir)
	watchEvents := make(chan watch.WatchEvent, 32)
	watchTopic.Subscribe(watchEvents)

	// We need watcher to calculate file hashes before we do initial build so we can detect changes
	// TODO: make watcher not wait for 1s before doing initial watch...
	select {
	case e := <-watchEvents:
		_, ok := e.(watch.WatchEventModuleAdded)
		if !ok {
			return fmt.Errorf("expected module added event, got: %T", e)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("expected module added event, got no event")
	case <-ctx.Done():
		return ctx.Err()
	}

	go func() {
		for {
			select {
			case e := <-watchEvents:
				log.FromContext(ctx).Infof("file changes e: %v", e)
				if _, ok := e.(watch.WatchEventModuleChanged); ok {
					log.FromContext(ctx).Infof("Found file changes: %s", buildCtx.Config.Dir)
					events <- filesUpdatedEvent{}
				}
				// TODO: remove
			// case <-time.After(1 * time.Second):
			// 	log.FromContext(ctx).Infof("idling")
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// TODO: explain
func buildContextFromPendingEvents(ctx context.Context, buildCtx buildContext, events chan updateEvent, firstEvent updateEvent) (newBuildCtx buildContext, isAutomaticRebuild bool) {
	allEvents := []updateEvent{firstEvent}
	// find any other events in the queue
	for {
		select {
		case e := <-events:
			allEvents = append(allEvents, e)
		case <-ctx.Done():
			return buildCtx, false
		default:
			log.FromContext(ctx).Infof("processing events: %v", allEvents)
			// No more events waiting to be processed
			hasExplicitBuilt := false
			for _, e := range allEvents {
				switch e := e.(type) {
				case buildContextUpdatedEvent:
					buildCtx = e.buildCtx
					hasExplicitBuilt = true
				case filesUpdatedEvent:
				}

			}
			switch e := firstEvent.(type) {
			case buildContextUpdatedEvent:
				buildCtx = e.buildCtx
				hasExplicitBuilt = true
			case filesUpdatedEvent:
			}
			return buildCtx, !hasExplicitBuilt
		}
	}
}

func buildAndSend(ctx context.Context, stream *connect.ServerStream[langpb.BuildEvent], projectRoot, stubsRoot string, buildCtx buildContext, isAutomaticRebuild bool, transaction compile.ModifyFilesTransaction) error {
	buildEvent, err := build(ctx, projectRoot, stubsRoot, buildCtx, isAutomaticRebuild, transaction)
	if err != nil {
		// wrap general errors into build events
		// TODO: how to send back non positional errors?
		buildEvent = &langpb.BuildEvent{
			Event: &langpb.BuildEvent_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					ContextId:          buildCtx.Id,
					IsAutomaticRebuild: isAutomaticRebuild,
					Errors: langpb.ErrorsToProto([]builderrors.Error{
						builderrors.Errorf(builderrors.Position{}, err.Error()),
					}),
					InvalidateDependencies: false,
				},
			},
		}
	}
	if err = stream.Send(buildEvent); err != nil {
		return fmt.Errorf("could not send build event: %w", err)
	}
	return nil
}

func build(ctx context.Context, projectRoot, stubsRoot string, buildCtx buildContext, isAutomaticRebuild bool, transaction compile.ModifyFilesTransaction) (*langpb.BuildEvent, error) {
	release, err := flock.Acquire(ctx, filepath.Join(buildCtx.Config.BuildLock), time.Second*1) //BuildLockTimeout)
	if err != nil {
		return nil, fmt.Errorf("could not acquire build lock: %w", err)
	}
	defer release() //nolint:errcheck

	deps, err := compile.ExtractDependencies(buildCtx.Config)
	if err != nil {
		return nil, err
	}

	if !slices.Equal(islices.Sort(deps), islices.Sort(buildCtx.Dependencies)) {
		// dependencies have changed
		return &langpb.BuildEvent{
			Event: &langpb.BuildEvent_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					ContextId:              buildCtx.Id,
					IsAutomaticRebuild:     isAutomaticRebuild,
					InvalidateDependencies: true,
				},
			},
		}, nil
	}

	// TODO: figure out buildEnv
	m, buildErrs, err := compile.Build(ctx, projectRoot, stubsRoot, buildCtx.Config, buildCtx.Schema, transaction, []string{}, false)
	if err != nil {
		return nil, err
	}
	module, ok := m.Get()
	if !ok {
		return &langpb.BuildEvent{
			Event: &langpb.BuildEvent_BuildFailure{
				BuildFailure: &langpb.BuildFailure{
					ContextId:              buildCtx.Id,
					IsAutomaticRebuild:     isAutomaticRebuild,
					Errors:                 langpb.ErrorsToProto(buildErrs),
					InvalidateDependencies: false,
				},
			},
		}, nil
	}
	moduleProto := module.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	return &langpb.BuildEvent{
		Event: &langpb.BuildEvent_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				ContextId:          buildCtx.Id,
				IsAutomaticRebuild: isAutomaticRebuild,
				Errors:             langpb.ErrorsToProto(buildErrs),
				Module:             moduleProto,
				Deploy:             []string{"main", "launch"},
			},
		},
	}, nil
}

func relativeWatchPatterns(moduleDir string, watchPaths []string) ([]string, error) {
	relativePaths := make([]string, len(watchPaths))
	for i, path := range watchPaths {
		relative, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return nil, fmt.Errorf("could create relative path for watch pattern: %w", err)
		}
		relativePaths[i] = relative
	}
	return relativePaths, nil
}
