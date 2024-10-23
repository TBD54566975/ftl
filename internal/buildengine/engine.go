package buildengine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine/languageplugin"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/ftl/internal/watch"
)

type schemaChange struct {
	ChangeType ftlv1.DeploymentChangeType
	*schema.Module
}

// moduleMeta is a wrapper around a module that includes the last build's start time.
type moduleMeta struct {
	module         Module
	plugin         languageplugin.LanguagePlugin
	events         chan languageplugin.PluginEvent
	configDefaults moduleconfig.CustomDefaults
}

// copyMetaWithUpdatedDependencies finds the dependencies for a module and returns a
// copy with those dependencies populated.
func copyMetaWithUpdatedDependencies(ctx context.Context, m moduleMeta) (moduleMeta, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting dependencies for %q", m.module.Config.Module)

	dependencies, err := m.plugin.GetDependencies(ctx, m.module.Config)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not get dependencies for %v: %w", m.module.Config.Module, err)
	}

	m.module = m.module.CopyWithDependencies(dependencies)
	return m, nil
}

// EngineEvent is an event published by the engine as modules get built and deployed.
//
//sumtype:decl
type EngineEvent interface {
	buildEvent()
}

// EngineStarted is published when the engine becomes busy building and deploying modules.
//
// For individual events as each module build starts, see ModuleBuildStarted
type EngineStarted struct{}

func (EngineStarted) buildEvent() {}

// EngineEnded is published when the engine is no longer building or deploying any modules.
// If there are any remaining errors, they will be included in the ModuleErrors map.
//
// For individual events as each module build ends, see ModuleBuildSuccess and ModuleBuildFailed
type EngineEnded struct {
	ModuleErrors map[string]error
}

func (EngineEnded) buildEvent() {}

// rawEngineEvent are events published from explicit builds and automatic rebuilds
// These are published to an internal chan for preprocessing before being published to the BuildUpdates topic
//
//sumtype:decl
type rawEngineEvent interface {
	rawBuildEvent()
}

// ModuleAdded is published when the engine discovers a module.
type ModuleAdded struct {
	Module string
}

func (ModuleAdded) buildEvent()    {}
func (ModuleAdded) rawBuildEvent() {}

// ModuleRemoved is published when the engine discovers a module has been removed.
type ModuleRemoved struct {
	Module string
}

func (ModuleRemoved) buildEvent()    {}
func (ModuleRemoved) rawBuildEvent() {}

// ModuleBuildWaiting is published when a build is waiting for dependencies to build
type ModuleBuildWaiting struct {
	Config moduleconfig.ModuleConfig
}

func (ModuleBuildWaiting) buildEvent()    {}
func (ModuleBuildWaiting) rawBuildEvent() {}

// ModuleBuildStarted is published when a build has started for a module.
type ModuleBuildStarted struct {
	Config        moduleconfig.ModuleConfig
	IsAutoRebuild bool
}

func (ModuleBuildStarted) buildEvent()    {}
func (ModuleBuildStarted) rawBuildEvent() {}

// ModuleBuildFailed is published for any build failures.
type ModuleBuildFailed struct {
	Config        moduleconfig.ModuleConfig
	Error         error
	IsAutoRebuild bool
}

func (ModuleBuildFailed) buildEvent()    {}
func (ModuleBuildFailed) rawBuildEvent() {}

// ModuleBuildSuccess is published when all modules have been built successfully built.
type ModuleBuildSuccess struct {
	Config        moduleconfig.ModuleConfig
	IsAutoRebuild bool
}

func (ModuleBuildSuccess) buildEvent()    {}
func (ModuleBuildSuccess) rawBuildEvent() {}

// ModuleDeployStarted is published when a deploy has begun for a module.
type ModuleDeployStarted struct {
	Module string
}

func (ModuleDeployStarted) buildEvent()    {}
func (ModuleDeployStarted) rawBuildEvent() {}

// ModuleDeployFailed is published for any deploy failures.
type ModuleDeployFailed struct {
	Module string
	Error  error
}

func (ModuleDeployFailed) buildEvent()    {}
func (ModuleDeployFailed) rawBuildEvent() {}

// ModuleDeploySuccess is published when all modules have been built successfully deployed.
type ModuleDeploySuccess struct {
	Module string
}

func (ModuleDeploySuccess) buildEvent()    {}
func (ModuleDeploySuccess) rawBuildEvent() {}

// invalidateDependenciesEvent is published when a module needs to be rebuilt when a module
// failed to buuld due to a change in dependencies.
type invalidateDependenciesEvent struct {
	module string
}

// Engine for building a set of modules.
type Engine struct {
	client           DeployClient
	bindAllocator    *bind.BindAllocator
	moduleMetas      *xsync.MapOf[string, moduleMeta]
	projectRoot      string
	moduleDirs       []string
	watcher          *watch.Watcher // only watches for module toml changes
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schemaChange]
	cancel           func()
	parallelism      int
	modulesToBuild   *xsync.MapOf[string, bool]
	buildEnv         []string
	devMode          bool
	startTime        optional.Option[time.Time]

	// events coming in from plugins
	pluginEvents chan languageplugin.PluginEvent

	invalidateDeps chan invalidateDependenciesEvent

	// internal channel for raw engine updates (does not include all state changes)
	rawEngineUpdates chan rawEngineEvent

	// topic to subscribe to engine events
	EngineUpdates *pubsub.Topic[EngineEvent]
}

type Option func(o *Engine)

func Parallelism(n int) Option {
	return func(o *Engine) {
		o.parallelism = n
	}
}

func BuildEnv(env []string) Option {
	return func(o *Engine) {
		o.buildEnv = env
	}
}

// WithDevMode sets the engine to dev mode.
func WithDevMode(devMode bool) Option {
	return func(o *Engine) {
		o.devMode = devMode
	}
}

// WithStartTime sets the start time to report total startup time
func WithStartTime(startTime time.Time) Option {
	return func(o *Engine) {
		o.startTime = optional.Some(startTime)
	}
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
//
// "dirs" are directories to scan for local modules.
func New(ctx context.Context, client DeployClient, projectRoot string, moduleDirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		projectRoot:      projectRoot,
		moduleDirs:       moduleDirs,
		moduleMetas:      xsync.NewMapOf[string, moduleMeta](),
		watcher:          watch.NewWatcher("ftl.toml"),
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schemaChange](),
		pluginEvents:     make(chan languageplugin.PluginEvent, 128),
		parallelism:      runtime.NumCPU(),
		modulesToBuild:   xsync.NewMapOf[string, bool](),
		invalidateDeps:   make(chan invalidateDependenciesEvent, 128),
		rawEngineUpdates: make(chan rawEngineEvent, 128),
		EngineUpdates:    pubsub.New[EngineEvent](),
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

	updateTerminalWithEngineEvents(ctx, e.EngineUpdates)

	go e.watchForPluginEvents(ctx)
	go e.watchForEventsToPublish(ctx)

	configs, err := watch.DiscoverModules(ctx, moduleDirs)
	if err != nil {
		return nil, fmt.Errorf("could not find modules: %w", err)
	}

	wg := &errgroup.Group{}
	for _, config := range configs {
		wg.Go(func() error {
			meta, err := e.newModuleMeta(ctx, config)
			if err != nil {
				return err
			}
			meta, err = copyMetaWithUpdatedDependencies(ctx, meta)
			if err != nil {
				return err
			}
			e.moduleMetas.Store(config.Module, meta)
			e.modulesToBuild.Store(config.Module, true)
			e.rawEngineUpdates <- ModuleAdded{Module: config.Module}
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return nil, err //nolint:wrapcheck
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
		deps = meta.module.Dependencies(AlwaysIncludeBuiltin)
	}
	if !foundModule {
		if sch, ok := e.controllerSchema.Load(moduleName); ok {
			foundModule = true
			deps = append(deps, sch.Imports()...)
		}
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
				meta, ok := e.moduleMetas.Load(moduleName)
				if !ok {
					return fmt.Errorf("module %q not found", moduleName)
				}
				if len(meta.module.Deploy) == 0 {
					return fmt.Errorf("no files found to deploy for %q", moduleName)
				}
				e.rawEngineUpdates <- ModuleDeployStarted{Module: moduleName}
				err := Deploy(ctx, meta.module, meta.module.Deploy, replicas, waitForDeployOnline, e.client)
				if err != nil {
					e.rawEngineUpdates <- ModuleDeployFailed{Module: moduleName, Error: err}
					return err
				}
				e.rawEngineUpdates <- ModuleDeploySuccess{Module: moduleName}
				return nil
			})
		}
		if err := deployGroup.Wait(); err != nil {
			return fmt.Errorf("deploy failed: %w", err)
		}
	}
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

// watchForModuleChanges watches for changes and all build start and event state changes.
func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	schemaChanges := make(chan schemaChange, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer func() {
		e.schemaChanges.Unsubscribe(schemaChanges)
	}()

	watchEvents := make(chan watch.WatchEvent, 128)
	ctx, cancel := context.WithCancel(ctx)
	topic, err := e.watcher.Watch(ctx, period, e.moduleDirs)
	if err != nil {
		cancel()
		return err
	}
	topic.Subscribe(watchEvents)
	defer func() {
		// Cancel will close the topic and channel
		cancel()
	}()

	// Build and deploy all modules first.
	err = e.BuildAndDeploy(ctx, 1, true)
	if err != nil {
		logger.Errorf(err, "initial deploy failed")
	}

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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event := <-watchEvents:
			switch event := event.(type) {
			case watch.WatchEventModuleAdded:
				config := event.Config
				if _, exists := e.moduleMetas.Load(config.Module); !exists {
					meta, err := e.newModuleMeta(ctx, config)
					if err != nil {
						logger.Errorf(err, "could not add module %s", config.Module)
						continue
					}
					e.moduleMetas.Store(config.Module, meta)
					e.rawEngineUpdates <- ModuleAdded{Module: config.Module}
					_ = e.BuildAndDeploy(ctx, 1, true, config.Module) //nolint:errcheck
				}
			case watch.WatchEventModuleRemoved:
				err := terminateModuleDeployment(ctx, e.client, event.Config.Module)
				if err != nil {
					logger.Errorf(err, "terminate %s failed", event.Config.Module)
				}
				if meta, ok := e.moduleMetas.Load(event.Config.Module); ok {
					meta.plugin.Updates().Unsubscribe(meta.events)
					err := meta.plugin.Kill()
					if err != nil {
						logger.Errorf(err, "terminate %s plugin failed", event.Config.Module)
					}
				}
				e.moduleMetas.Delete(event.Config.Module)
				e.rawEngineUpdates <- ModuleRemoved{Module: event.Config.Module}
			case watch.WatchEventModuleChanged:
				// ftl.toml file has changed
				meta, ok := e.moduleMetas.Load(event.Config.Module)
				if !ok {
					logger.Warnf("Module %q not found", event.Config.Module)
					continue
				}

				updatedConfig, err := moduleconfig.LoadConfig(event.Config.Dir)
				if err != nil {
					logger.Errorf(err, "Could not load updated toml for %s", event.Config.Module)
					continue
				}
				validConfig, err := updatedConfig.FillDefaultsAndValidate(meta.configDefaults)
				if err != nil {
					logger.Errorf(err, "Could not configure module config defaults for %s", event.Config.Module)
					continue
				}
				meta.module.Config = validConfig
				e.moduleMetas.Store(event.Config.Module, meta)

				_ = e.BuildAndDeploy(ctx, 1, true, event.Config.Module) //nolint:errcheck
			}
		case change := <-schemaChanges:
			if change.ChangeType == ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED {
				continue
			}
			existingHash, ok := moduleHashes[change.Name]
			if !ok {
				existingHash = []byte{}
			}

			hash, err := computeModuleHash(change.Module)
			if err != nil {
				logger.Errorf(err, "compute hash for %s failed", change.Name)
				continue
			}

			if bytes.Equal(hash, existingHash) {
				logger.Tracef("schema for %s has not changed", change.Name)
				continue
			}

			moduleHashes[change.Name] = hash

			dependentModuleNames := e.getDependentModuleNames(change.Name)
			if len(dependentModuleNames) > 0 {
				logger.Infof("%s's schema changed; processing %s", change.Name, strings.Join(dependentModuleNames, ", "))
				_ = e.BuildAndDeploy(ctx, 1, true, dependentModuleNames...) //nolint:errcheck
			}

		case event := <-e.invalidateDeps:
			_ = e.BuildAndDeploy(ctx, 1, true, event.module) //nolint:errcheck
		}
	}
}

// watchForEventsToPublish listens for raw build events, collects state, and publishes public events to BuildUpdates topic.
func (e *Engine) watchForEventsToPublish(ctx context.Context) {
	moduleErrors := map[string]error{}
	explicitlyBuilding := map[string]bool{}
	autoRebuilding := map[string]bool{}
	deploying := map[string]bool{}

	isIdle := true
	var endTime time.Time
	var becomeIdleTimer <-chan time.Time

	isFirstRound := true

	for {
		select {
		case <-ctx.Done():
			return

		case <-becomeIdleTimer:
			becomeIdleTimer = nil
			if len(explicitlyBuilding) > 0 || len(autoRebuilding) > 0 || len(deploying) > 0 {
				continue
			}
			isIdle = true

			if e.devMode && isFirstRound {
				logger := log.FromContext(ctx)
				if len(moduleErrors) > 0 {
					logger.Errorf(errors.Join(maps.Values(moduleErrors)...), "Initial build failed")
				} else if start, ok := e.startTime.Get(); ok {
					e.startTime = optional.None[time.Time]()
					logger.Infof("All modules deployed in %.2fs, watching for changes...", endTime.Sub(start).Seconds())
				} else {
					logger.Infof("All modules deployed, watching for changes...")
				}
			}
			isFirstRound = false

			publicBuildErrors := map[string]error{}
			maps.Copy(moduleErrors, publicBuildErrors)
			e.EngineUpdates.Publish(EngineEnded{ModuleErrors: publicBuildErrors})

		case rawEvent := <-e.rawEngineUpdates:
			switch event := rawEvent.(type) {

			case ModuleAdded:
				e.EngineUpdates.Publish(event)
			case ModuleRemoved:
				delete(moduleErrors, event.Module)
				delete(explicitlyBuilding, event.Module)
				delete(autoRebuilding, event.Module)
			case ModuleBuildWaiting:

			case ModuleBuildStarted:
				if isIdle {
					isIdle = false
					e.EngineUpdates.Publish(EngineStarted{})
				}
				if event.IsAutoRebuild {
					autoRebuilding[event.Config.Module] = true
				} else {
					explicitlyBuilding[event.Config.Module] = true
				}
				delete(moduleErrors, event.Config.Module)
				log.FromContext(ctx).Module(event.Config.Module).Scope("build").Infof("Building module")
			case ModuleBuildFailed:
				if event.IsAutoRebuild {
					delete(autoRebuilding, event.Config.Module)
				} else {
					delete(explicitlyBuilding, event.Config.Module)
				}
				moduleErrors[event.Config.Module] = event.Error
				log.FromContext(ctx).Module(event.Config.Module).Scope("build").Errorf(event.Error, "Build failed")
			case ModuleBuildSuccess:
				if event.IsAutoRebuild {
					delete(autoRebuilding, event.Config.Module)
				} else {
					delete(explicitlyBuilding, event.Config.Module)
				}
				delete(moduleErrors, event.Config.Module)
			case ModuleDeployStarted:
				if isIdle {
					isIdle = false
					e.EngineUpdates.Publish(EngineStarted{})
				}
				deploying[event.Module] = true
				delete(moduleErrors, event.Module)
			case ModuleDeployFailed:
				delete(deploying, event.Module)
				moduleErrors[event.Module] = event.Error
			case ModuleDeploySuccess:
				delete(deploying, event.Module)
				delete(moduleErrors, event.Module)
			}
			engineEvent, ok := rawEvent.(EngineEvent)
			if !ok {
				panic(fmt.Sprintf("unexpected raw event type: %T", rawEvent))
			}
			e.EngineUpdates.Publish(engineEvent)
		}
		if !isIdle && len(explicitlyBuilding) == 0 && len(autoRebuilding) == 0 && len(deploying) == 0 {
			endTime = time.Now()
			becomeIdleTimer = time.After(time.Second * 2)
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
		for _, dep := range meta.module.Dependencies(AlwaysIncludeBuiltin) {
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
				e.rawEngineUpdates <- ModuleDeployStarted{Module: module.Config.Module}
				err := Deploy(buildCtx, module, module.Deploy, replicas, waitForDeployOnline, e.client)
				if err != nil {
					e.rawEngineUpdates <- ModuleDeployFailed{Module: module.Config.Module, Error: err}
					return err
				}
				e.rawEngineUpdates <- ModuleDeploySuccess{Module: module.Config.Module}
				return nil
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
	if len(moduleNames) == 0 {
		e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
			moduleNames = append(moduleNames, name)
			return true
		})
	}

	mustBuildChan := make(chan moduleconfig.ModuleConfig, len(moduleNames))
	wg := errgroup.Group{}
	for _, name := range moduleNames {
		wg.Go(func() error {
			meta, ok := e.moduleMetas.Load(name)
			if !ok {
				return fmt.Errorf("module %q not found", name)
			}

			meta, err := copyMetaWithUpdatedDependencies(ctx, meta)
			if err != nil {
				return fmt.Errorf("could not get dependencies for %s: %w", name, err)
			}

			e.moduleMetas.Store(name, meta)
			mustBuildChan <- meta.module.Config
			return nil
		})
	}
	if err := wg.Wait(); err != nil {
		return err //nolint:wrapcheck
	}
	close(mustBuildChan)
	mustBuild := map[string]bool{}
	for config := range mustBuildChan {
		mustBuild[config.Module] = true
		e.rawEngineUpdates <- ModuleBuildWaiting{Config: config}
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

		metasMap := map[string]moduleMeta{}
		e.moduleMetas.Range(func(name string, meta moduleMeta) bool {
			metasMap[name] = meta
			return true
		})
		err = GenerateStubs(ctx, e.projectRoot, maps.Values(knownSchemas), metasMap)
		if err != nil {
			return err
		}

		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))

		wg := errgroup.Group{}
		wg.SetLimit(e.parallelism)
		for _, moduleName := range group {
			wg.Go(func() error {
				logger := log.FromContext(ctx).Module(moduleName).Scope("build")
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
		err = SyncStubReferences(ctx, e.projectRoot, moduleNames, metasMap)
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
		return fmt.Errorf("module %q not found", moduleName)
	}

	for _, dep := range meta.module.Dependencies(Raw) {
		if _, ok := builtModules[dep]; !ok {
			logger.Warnf("build skipped because dependency %q failed to build", dep)
			return nil
		}
	}

	e.rawEngineUpdates <- ModuleBuildStarted{Config: meta.module.Config}
	err := e.build(ctx, moduleName, builtModules, schemas)
	if err != nil {
		e.rawEngineUpdates <- ModuleBuildFailed{Config: meta.module.Config, Error: err}
	} else {
		e.rawEngineUpdates <- ModuleBuildSuccess{Config: meta.module.Config}
	}
	if err == nil && callback != nil {
		// load latest meta as it may have been updated
		meta, ok = e.moduleMetas.Load(moduleName)
		if !ok {
			return fmt.Errorf("module %q not found", moduleName)
		}
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

	moduleSchema, deploy, err := build(ctx, meta.plugin, e.projectRoot, languageplugin.BuildContext{
		Config:       meta.module.Config,
		Schema:       sch,
		Dependencies: meta.module.Dependencies(Raw),
	}, e.buildEnv, e.devMode)
	if err != nil {
		if errors.Is(err, errInvalidateDependencies) {
			// Do not start a build directly as we are already building out a graph of modules.
			// Instead we send to a chan so that it can be processed after.
			e.invalidateDeps <- invalidateDependenciesEvent{module: moduleName}
		}
		return err
	}
	// update files to deploy
	e.moduleMetas.Compute(moduleName, func(meta moduleMeta, exists bool) (out moduleMeta, shouldDelete bool) {
		if !exists {
			return moduleMeta{}, true
		}
		meta.module = meta.module.CopyWithDeploy(deploy)
		return meta, false
	})
	schemas <- moduleSchema
	return nil
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
		} else {
			// We don't want to use a remote schema if we have it locally
			delete(out, name)
		}
		return true
	})

	return nil
}

func (e *Engine) newModuleMeta(ctx context.Context, config moduleconfig.UnvalidatedModuleConfig) (moduleMeta, error) {
	plugin, err := languageplugin.New(ctx, e.bindAllocator, config.Language)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not create plugin for %s: %w", config.Module, err)
	}
	events := make(chan languageplugin.PluginEvent, 64)
	plugin.Updates().Subscribe(events)

	// pass on plugin events to the main event channel
	// make sure we do not pass on nil (chan closure) events
	go func() {
		for {
			select {
			case event := <-events:
				if event == nil {
					// chan closed
					return
				}
				e.pluginEvents <- event
			case <-ctx.Done():
				return
			}
		}
	}()

	// update config with defaults
	customDefaults, err := plugin.ModuleConfigDefaults(ctx, config.Dir)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not get defaults provider for %s: %w", config.Module, err)
	}
	validConfig, err := config.FillDefaultsAndValidate(customDefaults)
	if err != nil {
		return moduleMeta{}, fmt.Errorf("could not apply defaults for %s: %w", config.Module, err)
	}
	return moduleMeta{
		module:         newModule(validConfig),
		plugin:         plugin,
		events:         events,
		configDefaults: customDefaults,
	}, nil
}

// watchForPluginEvents listens for build updates from language plugins and reports them to the listener.
// These happen when a plugin for a module detects a change and automatically rebuilds.
func (e *Engine) watchForPluginEvents(originalCtx context.Context) {
	for {
		select {
		case event := <-e.pluginEvents:
			logger := log.FromContext(originalCtx).Module(event.ModuleName()).Scope("build")
			ctx := log.ContextWithLogger(originalCtx, logger)
			meta, ok := e.moduleMetas.Load(event.ModuleName())
			if !ok {
				logger.Warnf("module not found for build update")
				continue
			}
			switch event := event.(type) {
			case languageplugin.AutoRebuildStartedEvent:
				e.rawEngineUpdates <- ModuleBuildStarted{Config: meta.module.Config, IsAutoRebuild: true}

			case languageplugin.AutoRebuildEndedEvent:
				_, deploy, err := handleBuildResult(ctx, meta.module.Config, event.Result)
				if err != nil {
					e.rawEngineUpdates <- ModuleBuildFailed{Config: meta.module.Config, IsAutoRebuild: true, Error: err}
					if errors.Is(err, errInvalidateDependencies) {
						// Do not block this goroutine by building a module here.
						// Instead we send to a chan so that it can be processed elsewhere.
						e.invalidateDeps <- invalidateDependenciesEvent{module: event.ModuleName()}
					}
					continue
				}
				e.rawEngineUpdates <- ModuleBuildSuccess{Config: meta.module.Config, IsAutoRebuild: true}

				e.rawEngineUpdates <- ModuleDeployStarted{Module: event.Module}
				if err := Deploy(ctx, meta.module, deploy, 1, true, e.client); err != nil {
					e.rawEngineUpdates <- ModuleDeployFailed{Module: event.Module, Error: err}
					continue
				}
				e.rawEngineUpdates <- ModuleDeploySuccess{Module: event.Module}
			}
		case <-originalCtx.Done():
			return
		}
	}
}
