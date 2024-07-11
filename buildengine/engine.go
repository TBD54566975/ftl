package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
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
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/slices"
)

type CompilerBuildError struct {
	err error
}

func (e CompilerBuildError) Error() string {
	return e.err.Error()
}

func (e CompilerBuildError) Unwrap() error {
	return e.err
}

type schemaChange struct {
	ChangeType ftlv1.DeploymentChangeType
	*schema.Module
}

// moduleMeta is a wrapper around a module that includes the last build's start time.
type moduleMeta struct {
	module             Module
	lastBuildStartTime time.Time
}

type Listener interface {
	// OnBuildStarted is called when a build is started for a project.
	OnBuildStarted(module Module)

	// OnBuildSuccess is called when all modules have been built successfully and deployed.
	OnBuildSuccess()

	// OnBuildFailed is called for any build failures.
	// OnBuildSuccess should not be called if this is called after a OnBuildStarted.
	OnBuildFailed(err error)
}

// Engine for building a set of modules.
type Engine struct {
	client           ftlv1connect.ControllerServiceClient
	moduleMetas      *xsync.MapOf[string, moduleMeta]
	projectRoot      string
	moduleDirs       []string
	watcher          *Watcher
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schemaChange]
	cancel           func()
	parallelism      int
	listener         Listener
	modulesToBuild   *xsync.MapOf[string, bool]
}

type Option func(o *Engine)

func Parallelism(n int) Option {
	return func(o *Engine) {
		o.parallelism = n
	}
}

// WithListener sets the event listener for the Engine.
func WithListener(listener Listener) Option {
	return func(o *Engine) {
		o.listener = listener
	}
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
//
// "dirs" are directories to scan for local modules.
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient, projectRoot string, moduleDirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		projectRoot:      projectRoot,
		moduleDirs:       moduleDirs,
		moduleMetas:      xsync.NewMapOf[string, moduleMeta](),
		watcher:          NewWatcher(),
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schemaChange](),
		parallelism:      runtime.NumCPU(),
		modulesToBuild:   xsync.NewMapOf[string, bool](),
	}
	for _, option := range options {
		option(e)
	}
	e.controllerSchema.Store("builtin", schema.Builtins())
	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel

	err := CleanStubs(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to clean stubs: %w", err)
	}

	modules, err := DiscoverModules(ctx, moduleDirs)
	if err != nil {
		return nil, fmt.Errorf("could not find modules: %w", err)
	}
	for _, module := range modules {
		module, err = UpdateDependencies(ctx, module)
		if err != nil {
			return nil, err
		}
		e.moduleMetas.Store(module.Config.Module, moduleMeta{module: module})
		e.modulesToBuild.Store(module.Config.Module, true)
	}
	if client == nil {
		return e, nil
	}
	schemaSync := e.startSchemaSync(ctx)
	go rpc.RetryStreamingServerStream(ctx, backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, schemaSync, rpc.AlwaysRetry())
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
func (e *Engine) Graph(moduleNames ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(moduleNames) == 0 {
		e.moduleMetas.Range(func(name string, _ moduleMeta) bool {
			moduleNames = append(moduleNames, name)
			return true
		})
	}
	for _, name := range moduleNames {
		if err := e.buildGraph(name, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Engine) buildGraph(moduleName string, out map[string][]string) error {
	var deps []string
	// Short-circuit previously explored nodes
	if _, ok := out[moduleName]; ok {
		return nil
	}
	foundModule := false
	if meta, ok := e.moduleMetas.Load(moduleName); ok {
		foundModule = true
		deps = meta.module.Dependencies
	}
	if sch, ok := e.controllerSchema.Load(moduleName); ok {
		foundModule = true
		deps = append(deps, sch.Imports()...)
	}
	if !foundModule {
		return fmt.Errorf("module %q not found", moduleName)
	}
	deps = slices.Unique(deps)
	out[moduleName] = deps
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

// Each iterates over all local modules.
func (e *Engine) Each(fn func(Module) error) (err error) {
	e.moduleMetas.Range(func(key string, value moduleMeta) bool {
		if ferr := fn(value.module); ferr != nil {
			err = fmt.Errorf("%s: %w", key, ferr)
			return false
		}
		return true
	})
	return
}

// Deploy attempts to deploy all (already compiled) local modules.
//
// If waitForDeployOnline is true, this function will block until all deployments are online.
func (e *Engine) Deploy(ctx context.Context, replicas int32, waitForDeployOnline bool) error {
	graph, err := e.Graph(e.Modules()...)
	if err != nil {
		return err
	}

	groups, err := TopologicalSort(graph)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	for _, group := range groups {
		deployGroup, ctx := errgroup.WithContext(ctx)
		for _, moduleName := range group {
			if moduleName == "builtin" {
				continue
			}
			deployGroup.Go(func() error {
				module, ok := e.moduleMetas.Load(moduleName)
				if !ok {
					return fmt.Errorf("module %q not found", moduleName)
				}
				return Deploy(ctx, module.module, replicas, waitForDeployOnline, e.client)
			})
		}
		if err := deployGroup.Wait(); err != nil {
			return fmt.Errorf("deploy failed: %w", err)
		}
	}
	log.FromContext(ctx).Infof("All modules deployed")
	return nil
}

// Modules returns the names of all modules.
func (e *Engine) Modules() []string {
	var moduleNames []string
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		moduleNames = append(moduleNames, name)
		return true
	})
	return moduleNames
}

// Dev builds and deploys all local modules and watches for changes, redeploying as necessary.
func (e *Engine) Dev(ctx context.Context, period time.Duration) error {
	return e.watchForModuleChanges(ctx, period)
}

func (e *Engine) reportBuildFailed(err error) {
	if e.listener != nil {
		e.listener.OnBuildFailed(err)
	}
}

func (e *Engine) reportSuccess() {
	if e.listener != nil {
		e.listener.OnBuildSuccess()
	}
}

func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	schemaChanges := make(chan schemaChange, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer func() {
		e.schemaChanges.Unsubscribe(schemaChanges)
		close(schemaChanges)
	}()

	watchEvents := make(chan WatchEvent, 128)
	topic, err := e.watcher.Watch(ctx, period, e.moduleDirs)
	if err != nil {
		return err
	}
	topic.Subscribe(watchEvents)
	defer func() {
		topic.Unsubscribe(watchEvents)
		topic.Close()
		close(watchEvents)
	}()

	// Build and deploy all modules first.
	err = e.BuildAndDeploy(ctx, 1, true)
	if err != nil {
		logger.Errorf(err, "initial deploy failed")
		e.reportBuildFailed(err)
	} else {
		logger.Infof("All modules deployed, watching for changes...")
		e.reportSuccess()
	}

	moduleHashes := map[string][]byte{}
	e.controllerSchema.Range(func(name string, sch *schema.Module) bool {
		hash, err := computeModuleHash(sch)
		if err != nil {
			logger.Errorf(err, "compute hash for %s failed", name)
			e.reportBuildFailed(err)
			return false
		}
		moduleHashes[name] = hash
		return true
	})

	didUpdateDeployments := false
	// Track if there was an error, so that when deployments are complete we don't report success.
	didError := false
	// Watch for file and schema changes
	for {
		var completedUpdatesTimer <-chan time.Time
		if didUpdateDeployments {
			completedUpdatesTimer = time.After(period * 2)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-completedUpdatesTimer:
			logger.Infof("All modules deployed, watching for changes...")
			// Some cases, this will trigger after a build failure, so report accordingly.
			if !didError {
				e.reportSuccess()
			}

			didUpdateDeployments = false
		case event := <-watchEvents:
			switch event := event.(type) {
			case WatchEventModuleAdded:
				config := event.Module.Config
				if _, exists := e.moduleMetas.Load(config.Module); !exists {
					e.moduleMetas.Store(config.Module, moduleMeta{module: event.Module})
					didError = false
					err := e.BuildAndDeploy(ctx, 1, true, config.Module)
					if err != nil {
						didError = true
						e.reportBuildFailed(err)
						logger.Errorf(err, "deploy %s failed", config.Module)
					} else {
						didUpdateDeployments = true
					}
				}
			case WatchEventModuleRemoved:
				config := event.Module.Config

				err := terminateModuleDeployment(ctx, e.client, config.Module)
				if err != nil {
					didError = true
					e.reportBuildFailed(err)
					logger.Errorf(err, "terminate %s failed", config.Module)
				} else {
					didUpdateDeployments = true
				}

				e.moduleMetas.Delete(config.Module)
			case WatchEventModuleChanged:
				config := event.Module.Config

				meta, ok := e.moduleMetas.Load(config.Module)
				if !ok {
					logger.Warnf("module %q not found", config.Module)
					continue
				}

				if event.Time.Before(meta.lastBuildStartTime) {
					logger.Debugf("Skipping build and deploy; event time %v is before the last build time %v", event.Time, meta.lastBuildStartTime)
					continue // Skip this event as it's outdated
				}
				didError = false
				err := e.BuildAndDeploy(ctx, 1, true, config.Module)
				if err != nil {
					didError = true
					e.reportBuildFailed(err)
					logger.Errorf(err, "build and deploy failed for module %q", event.Module.Config.Module)
				} else {
					didUpdateDeployments = true
				}
			}
		case change := <-schemaChanges:
			if change.ChangeType != ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED {
				continue
			}

			hash, err := computeModuleHash(change.Module)
			if err != nil {
				didError = true
				e.reportBuildFailed(err)
				logger.Errorf(err, "compute hash for %s failed", change.Name)
				continue
			}

			if bytes.Equal(hash, moduleHashes[change.Name]) {
				logger.Tracef("schema for %s has not changed", change.Name)
				continue
			}

			moduleHashes[change.Name] = hash

			dependentModuleNames := e.getDependentModuleNames(change.Name)
			if len(dependentModuleNames) > 0 {
				logger.Infof("%s's schema changed; processing %s", change.Name, strings.Join(dependentModuleNames, ", "))
				didError = false
				err = e.BuildAndDeploy(ctx, 1, true, dependentModuleNames...)
				if err != nil {
					didError = true
					e.reportBuildFailed(err)
					logger.Errorf(err, "deploy %s failed", change.Name)
				} else {
					didUpdateDeployments = true
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

func (e *Engine) getDependentModuleNames(moduleName string) []string {
	dependentModuleNames := map[string]bool{}
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		for _, dep := range meta.module.Dependencies {
			if dep == moduleName {
				dependentModuleNames[name] = true
			}
		}
		return true
	})
	return maps.Keys(dependentModuleNames)
}

// BuildAndDeploy attempts to build and deploy all local modules.
func (e *Engine) BuildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, moduleNames ...string) error {
	logger := log.FromContext(ctx)
	if len(moduleNames) == 0 {
		moduleNames = e.Modules()
	}

	buildGroup := errgroup.Group{}

	buildGroup.Go(func() error {
		return e.buildWithCallback(ctx, func(buildCtx context.Context, module Module) error {
			buildGroup.Go(func() error {
				e.modulesToBuild.Store(module.Config.Module, false)
				return Deploy(buildCtx, module, replicas, waitForDeployOnline, e.client)
			})
			return nil
		}, moduleNames...)
	})

	// Wait for all build and deploy attempts to complete
	buildErr := buildGroup.Wait()

	pendingInitialBuilds := []string{}
	e.modulesToBuild.Range(func(name string, value bool) bool {
		if value {
			pendingInitialBuilds = append(pendingInitialBuilds, name)
		}
		return true
	})

	// Print out all modules that have yet to build if there are any errors
	if len(pendingInitialBuilds) > 0 {
		logger.Infof("Modules waiting to build: %s", strings.Join(pendingInitialBuilds, ", "))
	}

	return buildErr
}

type buildCallback func(ctx context.Context, module Module) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, moduleNames ...string) error {
	mustBuild := map[string]bool{}
	if len(moduleNames) == 0 {
		e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
			moduleNames = append(moduleNames, name)
			return true
		})
	}
	for _, name := range moduleNames {
		meta, ok := e.moduleMetas.Load(name)
		if !ok {
			return fmt.Errorf("module %q not found", name)
		}
		// Update dependencies before building.
		var err error
		module, err := UpdateDependencies(ctx, meta.module)
		if err != nil {
			return err
		}
		e.moduleMetas.Store(name, moduleMeta{module: module})
		mustBuild[name] = true
	}
	graph, err := e.Graph(moduleNames...)
	if err != nil {
		return err
	}
	builtModules := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	topology, err := TopologicalSort(graph)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1024)
	for _, group := range topology {
		knownSchemas := map[string]*schema.Module{}
		err := e.gatherSchemas(builtModules, knownSchemas)
		if err != nil {
			return err
		}

		metas := e.allModuleMetas()
		moduleConfigs := make([]moduleconfig.ModuleConfig, len(metas))
		for i, meta := range metas {
			moduleConfigs[i] = meta.module.Config
		}
		err = GenerateStubs(ctx, e.projectRoot, maps.Values(knownSchemas), moduleConfigs)
		if err != nil {
			return err
		}

		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))

		wg := errgroup.Group{}
		wg.SetLimit(e.parallelism)
		for _, moduleName := range group {
			wg.Go(func() error {
				logger := log.FromContext(ctx).Scope(moduleName)
				ctx := log.ContextWithLogger(ctx, logger)
				err := e.tryBuild(ctx, mustBuild, moduleName, builtModules, schemas, callback)
				if err != nil {
					errCh <- err
				}
				return nil
			})
		}

		err = wg.Wait()
		if err != nil {
			return err
		}

		// Now this group is built, collect all the schemas.
		close(schemas)
		for sch := range schemas {
			builtModules[sch.Name] = sch
		}

		moduleNames := []string{}
		for _, module := range knownSchemas {
			moduleNames = append(moduleNames, module.Name)
		}

		// Sync references to stubs if needed by the runtime
		err = SyncStubReferences(ctx, e.projectRoot, moduleNames, moduleConfigs)
		if err != nil {
			return err
		}
	}

	close(errCh)
	allErrors := []error{}
	for err := range errCh {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}

func (e *Engine) tryBuild(ctx context.Context, mustBuild map[string]bool, moduleName string, builtModules map[string]*schema.Module, schemas chan *schema.Module, callback buildCallback) error {
	logger := log.FromContext(ctx)

	if !mustBuild[moduleName] {
		return e.mustSchema(ctx, moduleName, builtModules, schemas)
	}

	meta, ok := e.moduleMetas.Load(moduleName)
	if !ok {
		return fmt.Errorf("Module %q not found", moduleName)
	}

	for _, dep := range meta.module.Dependencies {
		if _, ok := builtModules[dep]; !ok {
			logger.Warnf("build skipped because dependency %q failed to build", dep)
			return nil
		}
	}

	meta.lastBuildStartTime = time.Now()
	e.moduleMetas.Store(moduleName, meta)
	err := e.build(ctx, moduleName, builtModules, schemas)
	if err == nil && callback != nil {
		return callback(ctx, meta.module)
	}

	return err
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load(moduleName); ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, moduleName, builtModules, schemas)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, moduleName string, builtModules map[string]*schema.Module, schemas chan<- *schema.Module) error {
	meta, ok := e.moduleMetas.Load(moduleName)
	if !ok {
		return fmt.Errorf("module %q not found", moduleName)
	}

	sch := &schema.Schema{Modules: maps.Values(builtModules)}

	if e.listener != nil {
		e.listener.OnBuildStarted(meta.module)
	}

	err := Build(ctx, e.projectRoot, sch, meta.module, e.watcher.GetTransaction(meta.module.Config.Dir))
	if err != nil {
		return err
	}
	config := meta.module.Config
	moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(config.Dir, config.DeployDir, config.Schema))
	if err != nil {
		return fmt.Errorf("could not load schema for module %q: %w", config.Module, err)
	}
	schemas <- moduleSchema
	return nil
}

func (e *Engine) allModuleMetas() []moduleMeta {
	var out []moduleMeta
	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		out = append(out, meta)
		return true
	})
	return out
}

// Construct a combined schema for a module and its transitive dependencies.
func (e *Engine) gatherSchemas(
	moduleSchemas map[string]*schema.Module,
	out map[string]*schema.Module,
) error {
	e.controllerSchema.Range(func(name string, sch *schema.Module) bool {
		out[name] = sch
		return true
	})

	e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
		if _, ok := moduleSchemas[name]; ok {
			out[name] = moduleSchemas[name]
		}
		return true
	})

	return nil
}
