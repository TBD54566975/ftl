package common

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/pubsub"
	"github.com/beevik/etree"
	"github.com/go-viper/mapstructure/v2"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/TBD54566975/ftl"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	langpb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	langconnect "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
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
	ID           string
	Config       moduleconfig.AbsModuleConfig
	Schema       *schema.Schema
	Dependencies []string
}

func buildContextFromProto(proto *langpb.BuildContext) (buildContext, error) {
	sch, err := schema.FromProto(proto.Schema)
	if err != nil {
		return buildContext{}, fmt.Errorf("could not parse schema from proto: %w", err)
	}
	config := langpb.ModuleConfigFromProto(proto.ModuleConfig)
	return buildContext{
		ID:           proto.Id,
		Config:       config,
		Schema:       sch,
		Dependencies: proto.Dependencies,
	}, nil
}

type Service struct {
	updatesTopic          *pubsub.Topic[updateEvent]
	acceptsContextUpdates atomic.Value[bool]
	scaffoldFiles         *zip.Reader
}

var _ langconnect.LanguageServiceHandler = &Service{}

func New(scaffoldFiles *zip.Reader) *Service {
	return &Service{
		updatesTopic:  pubsub.New[updateEvent](),
		scaffoldFiles: scaffoldFiles,
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
				Name:    "group",
				Help:    "The Maven groupId of the project.",
				Default: ptr("com.example"),
			},
		},
	}), nil
}

// CreateModule generates files for a new module with the requested name
func (s *Service) CreateModule(ctx context.Context, req *connect.Request[langpb.CreateModuleRequest]) (*connect.Response[langpb.CreateModuleResponse], error) {
	logger := log.FromContext(ctx)
	projConfig := langpb.ProjectConfigFromProto(req.Msg.ProjectConfig)
	groupAny, ok := req.Msg.Flags.AsMap()["group"]
	if !ok {
		return nil, fmt.Errorf("group flag not set")
	}
	group, ok := groupAny.(string)
	if !ok {
		return nil, fmt.Errorf("group not a string")
	}

	packageDir := strings.ReplaceAll(group, ".", "/")
	opts := []scaffolder.Option{}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := struct {
		Dir        string
		Name       string
		Group      string
		PackageDir string
	}{
		Dir:        projConfig.Path,
		Name:       req.Msg.Name,
		Group:      group,
		PackageDir: packageDir,
	}
	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Dir)
	if err := internal.ScaffoldZip(s.scaffoldFiles, parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	return connect.NewResponse(&langpb.CreateModuleResponse{}), nil
}

func (s *Service) GenerateStubs(ctx context.Context, req *connect.Request[langpb.GenerateStubsRequest]) (*connect.Response[langpb.GenerateStubsResponse], error) {
	return connect.NewResponse(&langpb.GenerateStubsResponse{}), nil
}

func (s *Service) SyncStubReferences(ctx context.Context, req *connect.Request[langpb.SyncStubReferencesRequest]) (*connect.Response[langpb.SyncStubReferencesResponse], error) {
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

	// cancel context when stream ends so that watcher can be stopped
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
	if err := buildAndSend(ctx, stream, buildCtx, false); err != nil {
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
			if isAutomaticRebuild {
				err = stream.Send(&langpb.BuildEvent{
					Event: &langpb.BuildEvent_AutoRebuildStarted{
						AutoRebuildStarted: &langpb.AutoRebuildStarted{
							ContextId: buildCtx.ID,
						},
					},
				})
				if err != nil {
					return fmt.Errorf("could not send auto rebuild started event: %w", err)
				}
			}
			if err = buildAndSend(ctx, stream, buildCtx, isAutomaticRebuild); err != nil {
				return err
			}
		case <-ctx.Done():
			log.FromContext(ctx).Infof("Build call ending - ctx cancelled")
			return nil
		}
	}
}

func build(ctx context.Context, bctx buildContext) (*langpb.BuildEvent, error) {
	config := bctx.Config
	logger := log.FromContext(ctx)
	javaConfig, err := loadJavaConfig(config.LanguageConfig, config.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}
	if javaConfig.BuildTool == JavaBuildToolMaven {
		if err := setPOMProperties(ctx, config.Dir); err != nil {
			// This is not a critical error, things will probably work fine
			// TBH updating the pom is maybe not the best idea anyway
			logger.Warnf("unable to update ftl.version in %s: %s", config.Dir, err.Error())
		}
	}
	logger.Infof("Using build command '%s'", config.Build)
	command := exec.Command(ctx, log.Debug, config.Dir, "bash", "-c", config.Build)
	err = command.RunBuffered(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to build module %q: %w", config.Module, err)
	}

	buildErrs, err := loadProtoErrors(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load build errors: %w", err)
	}
	if builderrors.ContainsTerminalError(langpb.ErrorsFromProto(buildErrs)) {
		// skip reading schema
		return &langpb.BuildEvent{Event: &langpb.BuildEvent_BuildFailure{&langpb.BuildFailure{
			ContextId: bctx.ID,
			Errors:    buildErrs,
		}}}, nil
	}

	moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(config.DeployDir, "schema.pb"))
	if err != nil {
		return nil, fmt.Errorf("failed to read schema for module: %w", err)
	}

	moduleProto := moduleSchema.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	return &langpb.BuildEvent{
		Event: &langpb.BuildEvent_BuildSuccess{
			BuildSuccess: &langpb.BuildSuccess{
				ContextId: bctx.ID,
				Errors:    buildErrs,
				Module:    moduleProto,
				Deploy:    []string{"launch", "quarkus-app"},
			},
		},
	}, nil
}

// BuildContextUpdated is called whenever the build context is update while a Build call with "rebuild_automatically" is active.
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
	watchTopic, err := watcher.Watch(ctx, time.Second, []string{buildCtx.Config.Dir})
	if err != nil {
		return fmt.Errorf("could not watch for file changes: %w", err)
	}
	log.FromContext(ctx).Infof("Watching for file changes: %s", buildCtx.Config.Dir)
	watchEvents := make(chan watch.WatchEvent, 32)
	watchTopic.Subscribe(watchEvents)

	// We need watcher to calculate file hashes before we do initial build so we can detect changes
	select {
	case e := <-watchEvents:
		_, ok := e.(watch.WatchEventModuleAdded)
		if !ok {
			return fmt.Errorf("expected module added event, got: %T", e)
		}
	case <-time.After(3 * time.Second):
		return fmt.Errorf("expected module added event, got no event")
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	}

	go func() {
		for {
			select {
			case e := <-watchEvents:
				if _, ok := e.(watch.WatchEventModuleChanged); ok {
					log.FromContext(ctx).Infof("Found file changes: %s", buildCtx.Config.Dir)
					events <- filesUpdatedEvent{}
				}

			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// buildContextFromPendingEvents processes all pending events to determine the latest context and whether the build is automatic.
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

// buildAndSend builds the module and sends the build event to the stream.
//
// Build errors are sent over the stream as a BuildFailure event.
// This function only returns an error if events could not be send over the stream.
func buildAndSend(ctx context.Context, stream *connect.ServerStream[langpb.BuildEvent], buildCtx buildContext, isAutomaticRebuild bool) error {
	buildEvent, err := build(ctx, buildCtx)
	if err != nil {
		buildEvent = buildFailure(buildCtx, isAutomaticRebuild, builderrors.Error{
			Type:  builderrors.FTL,
			Level: builderrors.ERROR,
			Msg:   err.Error(),
		})
	}
	if err = stream.Send(buildEvent); err != nil {
		return fmt.Errorf("could not send build event: %w", err)
	}
	return nil
}

// buildFailure creates a BuildFailure event based on build errors.
func buildFailure(buildCtx buildContext, isAutomaticRebuild bool, errs ...builderrors.Error) *langpb.BuildEvent {
	return &langpb.BuildEvent{
		Event: &langpb.BuildEvent_BuildFailure{
			BuildFailure: &langpb.BuildFailure{
				ContextId:              buildCtx.ID,
				IsAutomaticRebuild:     isAutomaticRebuild,
				Errors:                 langpb.ErrorsToProto(errs),
				InvalidateDependencies: false,
			},
		},
	}
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

const JavaBuildToolMaven string = "maven"
const JavaBuildToolGradle string = "gradle"

type JavaConfig struct {
	BuildTool string `mapstructure:"build-tool"`
}

func loadJavaConfig(languageConfig any, language string) (JavaConfig, error) {
	var javaConfig JavaConfig
	err := mapstructure.Decode(languageConfig, &javaConfig)
	if err != nil {
		return JavaConfig{}, fmt.Errorf("failed to decode %s config: %w", language, err)
	}
	return javaConfig, nil
}

func (s *Service) ModuleConfigDefaults(ctx context.Context, req *connect.Request[langpb.ModuleConfigDefaultsRequest]) (*connect.Response[langpb.ModuleConfigDefaultsResponse], error) {
	defaults := langpb.ModuleConfigDefaultsResponse{
		GeneratedSchemaDir: ptr("src/main/ftl-module-schema"),
		// Watch defaults to files related to maven and gradle
		Watch:          []string{"pom.xml", "src/**", "build/generated", "target/generated-sources"},
		LanguageConfig: &structpb.Struct{Fields: map[string]*structpb.Value{}},
	}
	dir := req.Msg.Dir
	pom := filepath.Join(dir, "pom.xml")
	buildGradle := filepath.Join(dir, "build.gradle")
	buildGradleKts := filepath.Join(dir, "build.gradle.kts")
	if fileExists(pom) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolMaven)
		defaults.Build = ptr("mvn -B package")
		defaults.DeployDir = "target"
	} else if fileExists(buildGradle) || fileExists(buildGradleKts) {
		defaults.LanguageConfig.Fields["build-tool"] = structpb.NewStringValue(JavaBuildToolGradle)
		defaults.Build = ptr("gradle build")
		defaults.DeployDir = "build"
	} else {
		return nil, fmt.Errorf("could not find JVM build file in %s", dir)
	}

	return connect.NewResponse[langpb.ModuleConfigDefaultsResponse](&defaults), nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[langpb.DependenciesRequest]) (*connect.Response[langpb.DependenciesResponse], error) {
	dependencies := map[string]bool{}
	// We also attempt to look at kotlin files
	// As the Java module supports both
	kotin, kotlinErr := extractKotlinFTLImports(req.Msg.ModuleConfig.Name, req.Msg.ModuleConfig.Dir)
	if kotlinErr == nil {
		// We don't really care about the error case, its probably a Java project
		for _, imp := range kotin {
			dependencies[imp] = true
		}
	}
	javaImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(req.Msg.ModuleConfig.Dir, "src/main/java"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".java")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := javaImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == req.Msg.ModuleConfig.Name {
					continue
				}
				dependencies[module] = true
			}
		}
		return scanner.Err()
	})

	// We only error out if they both failed
	if err != nil && kotlinErr != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Java module: %w", req.Msg.ModuleConfig.Name, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return connect.NewResponse[langpb.DependenciesResponse](&langpb.DependenciesResponse{Modules: modules}), nil
}

func extractKotlinFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file while extracting dependencies: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == self {
					continue
				}
				dependencies[module] = true
			}
		}
		return scanner.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Kotlin module: %w", self, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

// setPOMProperties updates the ftl.version properties in the
// pom.xml file in the given base directory.
func setPOMProperties(ctx context.Context, baseDir string) error {
	logger := log.FromContext(ctx)
	ftlVersion := ftl.Version
	// If we are running in dev mode, ftl.Version will be "dev"
	// If we are running on CI it is a git sha
	if ftlVersion == "dev" || !strings.Contains(ftlVersion, ".") {
		ftlVersion = "1.0-SNAPSHOT"
	}

	pomFile := filepath.Clean(filepath.Join(baseDir, "pom.xml"))

	logger.Debugf("Setting ftl.version in %s to %s", pomFile, ftlVersion)

	tree := etree.NewDocument()
	if err := tree.ReadFromFile(pomFile); err != nil {
		return fmt.Errorf("unable to read %s: %w", pomFile, err)
	}
	root := tree.Root()

	parent := root.SelectElement("parent")
	versionSet := false
	if parent != nil {
		// You can't use properties in the parent
		// If they are using our parent then we want to update the version
		group := parent.SelectElement("groupId")
		artifact := parent.SelectElement("artifactId")
		if group.Text() == "xyz.block.ftl" && (artifact.Text() == "ftl-build-parent-java" || artifact.Text() == "ftl-build-parent-kotlin") {
			version := parent.SelectElement("version")
			if version != nil {
				version.SetText(ftlVersion)
				versionSet = true
			}
		}
	}

	err := updatePomProperties(root, pomFile, ftlVersion)
	if err != nil && !versionSet {
		// This is only a failure if we also did not update the parent
		return err
	}

	err = tree.WriteToFile(pomFile)
	if err != nil {
		return fmt.Errorf("unable to write %s: %w", pomFile, err)
	}
	return nil
}

func updatePomProperties(root *etree.Element, pomFile string, ftlVersion string) error {
	properties := root.SelectElement("properties")
	if properties == nil {
		return fmt.Errorf("unable to find <properties> in %s", pomFile)
	}
	version := properties.SelectElement("ftl.version")
	if version == nil {
		return fmt.Errorf("unable to find <properties>/<ftl.version> in %s", pomFile)
	}
	version.SetText(ftlVersion)
	return nil
}

func loadProtoErrors(config moduleconfig.AbsModuleConfig) (*langpb.ErrorList, error) {
	errorsPath := filepath.Join(config.DeployDir, "errors.pb")
	if _, err := os.Stat(errorsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	content, err := os.ReadFile(errorsPath)
	if err != nil {
		return nil, fmt.Errorf("could not load build errors file: %w", err)
	}

	errorspb := &langpb.ErrorList{}
	err = proto.Unmarshal(content, errorspb)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal build errors %w", err)
	}
	return errorspb, nil
}

func ptr(s string) *string {
	return &s
}
