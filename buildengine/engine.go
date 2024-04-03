package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type schemaChange struct {
	ChangeType ftlv1.DeploymentChangeType
	*schema.Module
}

// Engine for building a set of modules.
type Engine struct {
	client           ftlv1connect.ControllerServiceClient
	projects         map[ProjectKey]Project
	moduleDirs       []string
	externalDirs     []string
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schemaChange]
	cancel           func()
	parallelism      int
}

type Option func(o *Engine)

func Parallelism(n int) Option {
	return func(o *Engine) {
		o.parallelism = n
	}
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
//
// "dirs" are directories to scan for local modules.
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient, moduleDirs []string, externalDirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		moduleDirs:       moduleDirs,
		externalDirs:     externalDirs,
		projects:         map[ProjectKey]Project{},
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schemaChange](),
		parallelism:      runtime.NumCPU(),
	}
	for _, option := range options {
		option(e)
	}
	e.controllerSchema.Store("builtin", schema.Builtins())
	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	projects, err := DiscoverProjects(ctx, moduleDirs, externalDirs, true)
	if err != nil {
		return nil, fmt.Errorf("could not find projects: %w", err)
	}
	for _, project := range projects {
		project, err = UpdateDependencies(ctx, project)
		if err != nil {
			return nil, err
		}
		e.projects[project.Config().Key] = project
	}

	if client == nil {
		return e, nil
	}
	schemaSync := e.startSchemaSync(ctx)
	go rpc.RetryStreamingServerStream(ctx, backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, schemaSync)
	return e, nil
}

// Sync module schema changes from the FTL controller, as well as from manual
// updates, and merge them into a single schema map.
func (e *Engine) startSchemaSync(ctx context.Context) func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	logger := log.FromContext(ctx)
	// Blocking schema sync from the controller.
	psch, err := e.client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err == nil {
		sch, err := schema.FromProto(psch.Msg.Schema)
		if err == nil {
			for _, module := range sch.Modules {
				e.controllerSchema.Store(module.Name, module)
			}
		} else {
			logger.Debugf("Failed to parse schema from controller: %s", err)
		}
	} else {
		logger.Debugf("Failed to get schema from controller: %s", err)
	}

	// Sync module schema changes from the controller into the schema event source.
	return func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
		switch msg.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
			sch, err := schema.ModuleFromProto(msg.Schema)
			if err != nil {
				return err
			}
			e.controllerSchema.Store(sch.Name, sch)
			e.schemaChanges.Publish(schemaChange{ChangeType: msg.ChangeType, Module: sch})

		case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
			e.controllerSchema.Delete(msg.ModuleName)
			e.schemaChanges.Publish(schemaChange{ChangeType: msg.ChangeType, Module: nil})
		}
		return nil
	}
}

// Close stops the Engine's schema sync.
func (e *Engine) Close() error {
	e.cancel()
	return nil
}

// Graph returns the dependency graph for the given modules.
//
// If no modules are provided, the entire graph is returned. An error is returned if
// any dependencies are missing.
func (e *Engine) Graph(projects ...ProjectKey) (map[string][]string, error) {
	out := map[string][]string{}
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}
	for _, key := range projects {
		if err := e.buildGraph(string(key), out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Engine) buildGraph(key string, out map[string][]string) error {
	var deps []string
	if project, ok := e.projects[ProjectKey(key)]; ok {
		deps = project.Config().Dependencies
	} else if sch, ok := e.controllerSchema.Load(key); ok {
		deps = sch.Imports()
	} else {
		return fmt.Errorf("module %q not found", key)
	}
	out[key] = deps
	for _, dep := range deps {
		if err := e.buildGraph(dep, out); err != nil {
			return err
		}
	}
	return nil
}

// Import manually imports a schema for a module as if it were retrieved from
// the FTL controller.
func (e *Engine) Import(ctx context.Context, schema *schema.Module) {
	e.controllerSchema.Store(schema.Name, schema)
}

// Build attempts to build all local modules.
func (e *Engine) Build(ctx context.Context) error {
	return e.buildWithCallback(ctx, nil)
}

// Deploy attempts to build and deploy all local modules.
func (e *Engine) Deploy(ctx context.Context, replicas int32, waitForDeployOnline bool) error {
	return e.buildAndDeploy(ctx, replicas, waitForDeployOnline)
}

// Dev builds and deploys all local modules and watches for changes, redeploying as necessary.
func (e *Engine) Dev(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	// Build and deploy all modules first.
	err := e.buildAndDeploy(ctx, 1, true)
	if err != nil {
		logger.Errorf(err, "initial deploy failed")
	}

	logger.Infof("All modules deployed, watching for changes...")

	// Then watch for changes and redeploy.
	return e.watchForModuleChanges(ctx, period)
}

func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	moduleHashes := map[string][]byte{}
	e.controllerSchema.Range(func(name string, sch *schema.Module) bool {
		hash, err := computeModuleHash(sch)
		if err != nil {
			logger.Errorf(err, "compute hash for %s failed", name)
			return false
		}
		moduleHashes[name] = hash
		return true
	})

	schemaChanges := make(chan schemaChange, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer e.schemaChanges.Unsubscribe(schemaChanges)

	watchEvents := make(chan WatchEvent, 128)
	watch := Watch(ctx, period, e.moduleDirs, e.externalDirs)
	watch.Subscribe(watchEvents)
	defer watch.Unsubscribe(watchEvents)
	defer watch.Close()

	// Watch for file and schema changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watchEvents:
			switch event := event.(type) {
			case WatchEventProjectAdded:
				config := event.Project.Config()
				if _, exists := e.projects[config.Key]; !exists {
					e.projects[config.Key] = event.Project
					err := e.buildAndDeploy(ctx, 1, true, config.Key)
					if err != nil {
						logger.Errorf(err, "deploy %s failed", config.Key)
					}
				}
			case WatchEventProjectRemoved:
				config := event.Project.Config()
				if module, ok := event.Project.(Module); ok {
					err := teminateModuleDeployment(ctx, e.client, module.Module)
					if err != nil {
						logger.Errorf(err, "terminate %s failed", module.Module)
					}
				}
				delete(e.projects, config.Key)
			case WatchEventProjectChanged:
				config := event.Project.Config()
				err := e.buildAndDeploy(ctx, 1, true, config.Key)
				if err != nil {
					switch project := event.Project.(type) {
					case Module:
						logger.Errorf(err, "build and deploy failed for module %q: %v", project.Config().Key, err)
					case ExternalLibrary:
						logger.Errorf(err, "build failed for library %q: %v", project.Config().Key, err)
					}

				}
			}
		case change := <-schemaChanges:
			if change.ChangeType != ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED {
				continue
			}

			hash, err := computeModuleHash(change.Module)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", change.Name)
				continue
			}

			if bytes.Equal(hash, moduleHashes[change.Name]) {
				logger.Tracef("schema for %s has not changed", change.Name)
				continue
			}

			moduleHashes[change.Name] = hash

			dependentProjectKeys := e.getDependentProjectKeys(change.Name)
			if len(dependentProjectKeys) > 0 {
				//TODO: inaccurate log message for ext libs
				logger.Infof("%s's schema changed; processing %s", change.Name, strings.Join(StringsFromProjectKeys(dependentProjectKeys), ", "))
				err = e.buildAndDeploy(ctx, 1, true, dependentProjectKeys...)
				if err != nil {
					logger.Errorf(err, "deploy %s failed", change.Name)
				}
			}
		}

	}
}

func computeModuleHash(module *schema.Module) ([]byte, error) {
	hasher := sha256.New()
	data := []byte(module.String())
	if _, err := hasher.Write(data); err != nil {
		return nil, err // Handle errors that might occur during the write
	}

	return hasher.Sum(nil), nil
}

func (e *Engine) getDependentProjectKeys(name string) []ProjectKey {
	dependentProjectKeys := map[ProjectKey]bool{}
	for k, project := range e.projects {
		for _, dep := range project.Config().Dependencies {
			if dep == name {
				dependentProjectKeys[k] = true
			}
		}
	}
	return maps.Keys(dependentProjectKeys)
}

func (e *Engine) buildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, projects ...ProjectKey) error {
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}

	deployQueue := make(chan Project, len(projects))
	wg, ctx := errgroup.WithContext(ctx)

	// Build all projects and enqueue the modules for deployment.
	wg.Go(func() error {
		defer close(deployQueue)

		return e.buildWithCallback(ctx, func(ctx context.Context, project Project) error {
			if _, ok := project.(Module); ok {
				select {
				case deployQueue <- project:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}, projects...)
	})

	// Process deployment queue.
	for range len(projects) {
		wg.Go(func() error {
			for {
				select {
				case project, ok := <-deployQueue:
					if !ok {
						return nil
					}
					if module, ok := project.(Module); ok {
						err := Deploy(ctx, module, replicas, waitForDeployOnline, e.client)
						if err != nil {
							return err
						}
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}
	return wg.Wait()
}

type buildCallback func(ctx context.Context, project Project) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, projects ...ProjectKey) error {
	mustBuild := map[ProjectKey]bool{}
	if len(projects) == 0 {
		projects = maps.Keys(e.projects)
	}
	for _, key := range projects {
		project, ok := e.projects[key]
		if !ok {
			return fmt.Errorf("project %q not found", key)
		}
		// Update dependencies before building.
		var err error
		project, err = UpdateDependencies(ctx, project)
		if err != nil {
			return err
		}
		e.projects[key] = project
		mustBuild[key] = true
	}
	graph, err := e.Graph(projects...)
	if err != nil {
		return err
	}
	builtModules := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	topology := TopologicalSort(graph)
	for _, group := range topology {
		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))
		wg, ctx := errgroup.WithContext(ctx)
		wg.SetLimit(e.parallelism)
		for _, keyStr := range group {
			key := ProjectKey(keyStr)
			wg.Go(func() error {
				if !mustBuild[key] {
					return e.mustSchema(ctx, key, builtModules, schemas)
				}

				err := e.build(ctx, key, builtModules, schemas)
				if err == nil && callback != nil {
					return callback(ctx, e.projects[key])
				}
				return err
			})
		}
		err := wg.Wait()
		if err != nil {
			return err
		}
		// Now this group is built, collect all the schemas.
		close(schemas)
		for sch := range schemas {
			builtModules[sch.Name] = sch
		}
	}

	return nil
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, key ProjectKey, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load(string(key)); ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, key, builtModules, schemas)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, key ProjectKey, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	project, ok := e.projects[key]
	if !ok {
		return fmt.Errorf("project %q not found", key)
	}

	combined := map[string]*schema.Module{}
	if err := e.gatherSchemas(builtModules, project, combined); err != nil {
		return err
	}
	sch := &schema.Schema{Modules: maps.Values(combined)}

	err := Build(ctx, sch, project)
	if err != nil {
		return err
	}
	if module, ok := project.(Module); ok {
		moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(module.Dir, module.DeployDir, module.Schema))
		if err != nil {
			return fmt.Errorf("could not load schema for module %q: %w", module.Config().Key, err)
		}
		schemas <- moduleSchema
	}
	return nil
}

// Construct a combined schema for a project and its transitive dependencies.
func (e *Engine) gatherSchemas(
	moduleSchemas map[string]*schema.Module,
	project Project,
	out map[string]*schema.Module,
) error {
	latestModule, ok := e.projects[project.Config().Key]
	if !ok {
		latestModule = project
	}
	for _, dep := range latestModule.Config().Dependencies {
		out[dep] = moduleSchemas[dep]
		if dep != "builtin" {
			depModule, ok := e.projects[ProjectKey(dep)]
			// TODO: should we be gathering schemas from dependencies without a project?
			// This can happen if the schema is loaded from the controller
			if ok {
				if err := e.gatherSchemas(moduleSchemas, depModule, out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
