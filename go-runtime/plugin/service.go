package main

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	languagepb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulewatcher"
	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/types/pubsub"
)

type metadata struct {
	name string
}

//sumtype:decl
type updateEvent interface{ updateEvent() }

type schemaUpdatedEvent struct {
	schema *schema.Schema
}

func (schemaUpdatedEvent) updateEvent() {}

type MetadataUpdatedEvent struct {
	metadata *metadata
}

type filesUpdatedEvent struct{}

func (filesUpdatedEvent) updateEvent() {}

func (MetadataUpdatedEvent) updateEvent() {}

func metadataFromProto(p *languagepb.Metadata) *metadata {
	return &metadata{
		name: p.Name,
	}
}

type buildContext struct {
	path        string
	projectPath string

	metadata metadata
	// dependencies []string
	schema *schema.Schema
}

func buildContextFromRequest(req *languagepb.BuildRequest) (buildContext, error) {
	sch, err := schema.FromProto(req.Schema)
	if err != nil {
		return buildContext{}, err
	}

	md := metadata{
		name: req.Metadata.Name,
	}

	return buildContext{
		path:        req.Path,
		projectPath: req.ProjectPath,
		metadata:    md,
		schema:      sch,
	}, nil
}

func (s buildContext) processEvent(e updateEvent) buildContext {
	switch e := e.(type) {
	case schemaUpdatedEvent:
		s.schema = e.schema
	case MetadataUpdatedEvent:
		s.metadata = *e.metadata
	}
	return s
}

type Service struct {
	path string

	updatesTopic *pubsub.Topic[updateEvent]
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	log.FromContext(ctx).Infof("Received Ping")
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

// Generates files for a new empty module with the requested name
func (s *Service) CreateModule(ctx context.Context, req *connect.Request[languagepb.CreateModuleRequest]) (*connect.Response[languagepb.CreateModuleResponse], error) {
	logger := log.FromContext(ctx)
	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
		scaffolder.Functions(template.FuncMap{
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
		}),
	}
	// TODO: bring back this logic
	// if !includeBinDir {
	logger.Debugf("Excluding bin directory")
	opts = append(opts, scaffolder.Exclude("^bin"))
	// }
	sctx := scaffoldingContext{
		Name:      req.Msg.Name,
		GoVersion: runtime.Version()[2:],
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(req.Msg.Path)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return nil, fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy")
	if err := exec.Command(ctx, log.Debug, req.Msg.Path, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return nil, err
	}
	return connect.NewResponse(&languagepb.CreateModuleResponse{}), nil
}

// Extract dependencies for a module
func (s *Service) GetDependencies(ctx context.Context, req *connect.Request[languagepb.DependenciesRequest]) (*connect.Response[languagepb.DependenciesResponse], error) {
	deps, err := compile.ExtractDependencies(req.Msg.Metadata.Name, req.Msg.Path)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&languagepb.DependenciesResponse{
		Modules: deps,
	}), nil
}

func (s *Service) MetadataUpdated(ctx context.Context, req *connect.Request[languagepb.MetadataUpdatedRequest]) (*connect.Response[languagepb.MetadataUpdatedResponse], error) {
	s.updatesTopic.Publish(MetadataUpdatedEvent{
		metadata: metadataFromProto(req.Msg.Metadata),
	})
	return connect.NewResponse(&languagepb.MetadataUpdatedResponse{}), nil
}

// SchemaUpdated is called whenever the relevant part of a schema is updated.
func (s *Service) SchemaUpdated(ctx context.Context, req *connect.Request[languagepb.SchemaUpdatedRequest]) (*connect.Response[languagepb.SchemaUpdatedResponse], error) {
	sch, err := schema.FromProto(req.Msg.Schema)
	if err != nil {
		return nil, err
	}
	s.updatesTopic.Publish(schemaUpdatedEvent{
		schema: sch,
	})
	return connect.NewResponse(&languagepb.SchemaUpdatedResponse{}), nil
}

// Build the module
func (s *Service) Build(ctx context.Context, req *connect.Request[languagepb.BuildRequest], stream *connect.ServerStream[languagepb.BuildEvent]) error {
	events := make(chan updateEvent, 32)
	s.updatesTopic.Subscribe(events)
	defer s.updatesTopic.Unsubscribe(events)

	buildCtx, err := buildContextFromRequest(req.Msg)
	if err != nil {
		return err
	}

	// Load initial dependencies and send over the stream to avoid race conditions
	deps, err := compile.ExtractDependencies(req.Msg.Metadata.Name, req.Msg.Path)
	if err != nil {
		return err
	}
	stream.Send(&languagepb.BuildEvent{
		Event: &languagepb.BuildEvent_DependencyUpdate{
			DependencyUpdate: &languagepb.DependencyUpdate{
				Modules: deps,
			},
		},
	})

	if req.Msg.Watch {
		// TODO: we need the watcher so we can get file change transactions
		if err := watchFiles(ctx, buildCtx, events); err != nil {
			return err
		}
	}

	// Initial build
	if err := buildAndSend(ctx, stream, buildCtx); err != nil {
		return err
	}
	if !req.Msg.Watch {
		log.FromContext(ctx).Infof("Build call ending - not watching")
		return nil
	}

	// Watch for changes and build as needed
	for {
		select {
		case e := <-events:
			processAllPendingEvents(buildCtx, events, e)
			if err = buildAndSend(ctx, stream, buildCtx); err != nil {
				return err
			}
		case <-ctx.Done():
			log.FromContext(ctx).Infof("Build call ending - ctx cancelled")
			return nil
		}
	}
}

func watchFiles(ctx context.Context, buildCtx buildContext, events chan updateEvent) error {
	// TODO: add back patterns from module config via grpc
	// TODO: do we need to stop the watcher when the build call ends? or is that automatic?
	watcher := modulewatcher.New("**/*.go", "go.mod", "go.sum")
	watchTopic, err := watcher.Watch(ctx, time.Second, []string{buildCtx.path})
	if err != nil {
		return fmt.Errorf("could not watch for file changes: %w", err)
	}
	log.FromContext(ctx).Infof("Watching for file changes: %s", buildCtx.path)
	watchEvents := make(chan modulewatcher.WatchEvent, 32)
	watchTopic.Subscribe(watchEvents)
	go func() {
		for {
			select {
			case e := <-watchEvents:
				log.FromContext(ctx).Infof("file changes e: %v", e)
				if _, ok := e.(modulewatcher.WatchEventModuleChanged); ok {
					// TODO: ignore ftl.toml changes?
					log.FromContext(ctx).Infof("Found file changes: %s", buildCtx.path)
					events <- filesUpdatedEvent{}
				}
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
func processAllPendingEvents(buildCtx buildContext, events chan updateEvent, firstEvent updateEvent) buildContext {
	buildCtx = buildCtx.processEvent(firstEvent)
	// process any other events in the queue
	for {
		select {
		case e := <-events:
			buildCtx = buildCtx.processEvent(e)
		default:
			// No more events waiting to be processed
			return buildCtx
		}
	}
}

func buildAndSend(ctx context.Context,
	stream *connect.ServerStream[languagepb.BuildEvent],
	buildCtx buildContext) error {
	result := build(ctx, buildCtx)
	return stream.Send(&languagepb.BuildEvent{
		Event: &languagepb.BuildEvent_BuildResult{
			BuildResult: result,
		},
	})
}

func build(ctx context.Context, buildCtx buildContext) *languagepb.BuildResult {
	// TODO: figure out last 2 args...
	m, buildErrs, err := compile.Build(ctx, buildCtx.projectPath, buildCtx.path, buildCtx.schema, compile.DummyTransaction{}, []string{}, false)
	if err != nil {
		return &languagepb.BuildResult{
			Deploy: []string{},
			Errors: languagepb.ErrorsToProto([]*builderrors.Error{
				builderrors.Errorf(builderrors.Position{}, err.Error()),
			}),
		}
	}
	var moduleProto *schemapb.Module
	if module, ok := m.Get(); ok {
		moduleProto = module.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	}
	return &languagepb.BuildResult{
		Module: moduleProto,
		// paths for files/directories to be deployed
		Deploy: []string{
			".ftl/main",
		},
		// errors and warnings encountered during the build
		Errors: languagepb.ErrorsToProto(buildErrs),
	}
}
