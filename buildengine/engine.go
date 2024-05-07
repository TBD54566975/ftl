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
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type schemaChange struct {
	ChangeType ftlv1.DeploymentChangeType
	*schema.Module
}

type projectMeta struct {
	project            Project
	lastBuildStartTime time.Time
}

type Listener interface {
	OnBuildStarted(project Project)
}

type BuildStartedListenerFunc func(project Project)

func (b BuildStartedListenerFunc) OnBuildStarted(project Project) { b(project) }

// Engine for building a set of modules.
type Engine struct {
	client           ftlv1connect.ControllerServiceClient
	projectConfig    *projectconfig.Config
	projectMetas     *xsync.MapOf[ProjectKey, projectMeta]
	moduleDirs       []string
	externalDirs     []string
	watcher          *Watcher
	controllerSchema *xsync.MapOf[string, *schema.Module]
	schemaChanges    *pubsub.Topic[schemaChange]
	cancel           func()
	parallelism      int
	listener         Listener
	projectsToBuild  *xsync.MapOf[ProjectKey, bool]
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
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient, projConfig *projectconfig.Config, moduleDirs []string, externalDirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		projectConfig:    projConfig,
		moduleDirs:       moduleDirs,
		externalDirs:     externalDirs,
		projectMetas:     xsync.NewMapOf[ProjectKey, projectMeta](),
		watcher:          NewWatcher(),
		controllerSchema: xsync.NewMapOf[string, *schema.Module](),
		schemaChanges:    pubsub.New[schemaChange](),
		parallelism:      runtime.NumCPU(),
		projectsToBuild:  xsync.NewMapOf[ProjectKey, bool](),
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
		e.projectMetas.Store(project.Config().Key, projectMeta{project: project})
		e.projectsToBuild.Store(project.Config().Key, true)
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
		e.projectMetas.Range(func(key ProjectKey, _ projectMeta) bool {
			projects = append(projects, key)
			return true
		})
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
	if meta, ok := e.projectMetas.Load(ProjectKey(key)); ok {
		deps = meta.project.Config().Dependencies
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
func (e *Engine) Dev(ctx context.Context, period time.Duration, commands projectconfig.Commands) error {
	logger := log.FromContext(ctx)
	if len(commands.Startup) > 0 {
		for _, cmd := range commands.Startup {
			logger.Debugf("Executing startup command: %s", cmd)
			if err := exec.Command(ctx, log.Info, ".", "bash", "-c", cmd).Run(); err != nil {
				return fmt.Errorf("startup command failed: %w", err)
			}
		}
	}

	return e.watchForModuleChanges(ctx, period)
}

func (e *Engine) watchForModuleChanges(ctx context.Context, period time.Duration) error {
	logger := log.FromContext(ctx)

	schemaChanges := make(chan schemaChange, 128)
	e.schemaChanges.Subscribe(schemaChanges)
	defer e.schemaChanges.Unsubscribe(schemaChanges)

	watchEvents := make(chan WatchEvent, 128)
	topic, err := e.watcher.Watch(ctx, period, e.moduleDirs, e.externalDirs)
	if err != nil {
		return err
	}
	topic.Subscribe(watchEvents)
	defer topic.Unsubscribe(watchEvents)
	defer topic.Close()

	// Build and deploy all modules first.
	err = e.buildAndDeploy(ctx, 1, true)
	if err != nil {
		logger.Errorf(err, "initial deploy failed")
	} else {
		logger.Infof("All modules deployed, watching for changes...")
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

	// Watch for file and schema changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watchEvents:
			switch event := event.(type) {
			case WatchEventProjectAdded:
				config := event.Project.Config()
				if _, exists := e.projectMetas.Load(config.Key); !exists {
					e.projectMetas.Store(config.Key, projectMeta{project: event.Project})
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
				e.projectMetas.Delete(config.Key)
			case WatchEventProjectChanged:
				config := event.Project.Config()

				meta, ok := e.projectMetas.Load(config.Key)
				if !ok {
					logger.Warnf("project %q not found", config.Key)
					continue
				}

				if event.Time.Before(meta.lastBuildStartTime) {
					logger.Debugf("Skipping build and deploy; event time %v is before the last build time %v", event.Time, meta.lastBuildStartTime)
					continue // Skip this event as it's outdated
				}
				err := e.buildAndDeploy(ctx, 1, true, config.Key)
				if err != nil {
					switch project := event.Project.(type) {
					case Module:
						logger.Errorf(err, "build and deploy failed for module %q", project.Config().Key)
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
	e.projectMetas.Range(func(key ProjectKey, meta projectMeta) bool {
		for _, dep := range meta.project.Config().Dependencies {
			if dep == name {
				dependentProjectKeys[key] = true
			}
		}
		return true
	})
	return maps.Keys(dependentProjectKeys)
}

func (e *Engine) buildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, projects ...ProjectKey) error {
	logger := log.FromContext(ctx)
	if len(projects) == 0 {
		e.projectMetas.Range(func(key ProjectKey, meta projectMeta) bool {
			projects = append(projects, key)
			return true
		})
	}

	buildGroup := errgroup.Group{}
	deployGroup := errgroup.Group{}

	buildGroup.Go(func() error {
		return e.buildWithCallback(ctx, func(buildCtx context.Context, builtProject Project) error {
			deployGroup.Go(func() error {
				e.projectsToBuild.Store(builtProject.Config().Key, false)
				module, ok := builtProject.(Module)
				if !ok {
					// Skip deploying external libraries
					return nil
				}
				return Deploy(buildCtx, module, replicas, waitForDeployOnline, e.client)
			})
			return nil
		}, projects...)
	})

	// Wait for all build and deploy attempts to complete
	buildErr := buildGroup.Wait()
	deployErr := deployGroup.Wait()

	pendingInitialBuilds := []string{}
	e.projectsToBuild.Range(func(key ProjectKey, value bool) bool {
		if value {
			pendingInitialBuilds = append(pendingInitialBuilds, string(key))
		}
		return true
	})

	// Print out all modules that have yet to build if there are any errors
	if len(pendingInitialBuilds) > 0 {
		logger.Infof("Modules waiting to build: %s", strings.Join(pendingInitialBuilds, ", "))
	}

	if buildErr != nil {
		return buildErr
	}

	return deployErr
}

type buildCallback func(ctx context.Context, project Project) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, projects ...ProjectKey) error {
	mustBuild := map[ProjectKey]bool{}
	if len(projects) == 0 {
		e.projectMetas.Range(func(key ProjectKey, meta projectMeta) bool {
			projects = append(projects, key)
			return true
		})
	}
	for _, key := range projects {
		meta, ok := e.projectMetas.Load(key)
		if !ok {
			return fmt.Errorf("project %q not found", key)
		}
		// Update dependencies before building.
		var err error
		project, err := UpdateDependencies(ctx, meta.project)
		if err != nil {
			return err
		}
		e.projectMetas.Store(key, projectMeta{project: project})
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
	errCh := make(chan error, 1024)
	for _, group := range topology {
		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))

		wg := errgroup.Group{}
		wg.SetLimit(e.parallelism)
		for _, keyStr := range group {
			key := ProjectKey(keyStr)

			wg.Go(func() error {
				logger := log.FromContext(ctx).Scope(string(key))
				ctx := log.ContextWithLogger(ctx, logger)
				err := e.tryBuild(ctx, mustBuild, key, builtModules, schemas, callback)
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
	}

	close(errCh)
	allErrors := []error{}
	for err := range errCh {
		allErrors = append(allErrors, err)
	}

	allErrors = append(allErrors, e.validateConfigsAndSecretsMatch(ctx, builtModules)...)

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}

	return nil
}

func (e *Engine) validateConfigsAndSecretsMatch(ctx context.Context, builtModules map[string]*schema.Module) []error {
	errs := []error{}
	logger := log.FromContext(ctx)

	configsProvidedGlobally := make(map[string]bool)
	secretsProvidedGlobally := make(map[string]bool)
	if e.projectConfig != nil {
		for configName := range e.projectConfig.Global.Config {
			configsProvidedGlobally[configName] = true
		}
		for secretName := range e.projectConfig.Global.Secrets {
			secretsProvidedGlobally[secretName] = true
		}
	}

	configsUsed := make(map[string]bool)
	secretsUsed := make(map[string]bool)
	for moduleName, module := range builtModules {
		configsUsedInModule, secretsUsedInModule, moduleErrs := e.validateConfigsAndSecretsMatchForModule(ctx, moduleName, module, configsProvidedGlobally, secretsProvidedGlobally)
		errs = append(errs, moduleErrs...)
		for configName := range configsUsedInModule {
			configsUsed[configName] = true
		}
		for secretName := range secretsUsedInModule {
			secretsUsed[secretName] = true
		}
	}

	if e.projectConfig != nil {
		for configName := range e.projectConfig.Global.Config {
			if _, isUsed := configsUsed[configName]; !isUsed {
				logger.Warnf("config %q is provided globally in ftl-project.toml, but is not required by any modules", configName)
			}
		}
		for secretName := range e.projectConfig.Global.Secrets {
			if _, isUsed := secretsUsed[secretName]; !isUsed {
				logger.Warnf("secret %q is provided globally in ftl-project.toml, but is not required by any modules", secretName)
			}
		}
	}

	return errs
}

// validateConfigsAndSecretsMatchForModule is a helper function for validateConfigsAndSecretsMatch.
// `globalConfig` and `globalSecrets` store the names of all the configs/secrets defined globally in
// ftl-project.toml, with O(1) `contains` checks. This function logs warnings for any module-level
// configs/secrets that are provided but not used, then returns maps whose keys are all the
// configs/secrets used by this module.
func (e *Engine) validateConfigsAndSecretsMatchForModule(ctx context.Context, moduleName string, module *schema.Module, globalConfig map[string]bool, globalSecrets map[string]bool) (map[string]bool, map[string]bool, []error) {
	errs := []error{}
	logger := log.FromContext(ctx)

	configsUsed := make(map[string]bool)
	secretsUsed := make(map[string]bool)
	for _, d := range module.Decls {
		switch d := d.(type) {
		case *schema.Config:
			configsUsed[d.Name] = true
		case *schema.Secret:
			secretsUsed[d.Name] = true
		default:
		}
	}

	// Index all provided configs into configsProvided and warn for unused configs
	configsProvided := maps.Clone(globalConfig)
	secretsProvided := maps.Clone(globalSecrets)
	if e.projectConfig != nil {
		moduleConfigAndSecrets, moduleConfigAndSecretsExists := e.projectConfig.Modules[moduleName]
		if moduleConfigAndSecretsExists {
			for configName := range moduleConfigAndSecrets.Config {
				configsProvided[configName] = true
				if _, isUsed := configsUsed[configName]; !isUsed {
					logger.Warnf("config %q is provided for module %q in ftl-project.toml, but is not required", configName, moduleName)
				}
			}
			for secretName := range moduleConfigAndSecrets.Secrets {
				secretsProvided[secretName] = true
				if _, isUsed := secretsUsed[secretName]; !isUsed {
					logger.Warnf("secret %q is provided for module %q in ftl-project.toml, but is not required", secretName, moduleName)
				}
			}
		}
	}

	for configName := range configsUsed {
		if _, isProvided := configsProvided[configName]; !isProvided {
			errs = append(errs, fmt.Errorf("config %q is not provided in ftl-project.toml, but is required by module %q", configName, moduleName))
		}
	}
	for secretName := range secretsUsed {
		if _, isProvided := secretsProvided[secretName]; !isProvided {
			errs = append(errs, fmt.Errorf("secret %q is not provided in ftl-project.toml, but is required by module %q", secretName, moduleName))
		}
	}

	return configsUsed, secretsUsed, errs
}

func (e *Engine) tryBuild(ctx context.Context, mustBuild map[ProjectKey]bool, key ProjectKey, builtModules map[string]*schema.Module, schemas chan *schema.Module, callback buildCallback) error {
	logger := log.FromContext(ctx)

	if !mustBuild[key] {
		return e.mustSchema(ctx, key, builtModules, schemas)
	}

	meta, ok := e.projectMetas.Load(key)
	if !ok {
		return fmt.Errorf("project %q not found", key)
	}

	for _, dep := range meta.project.Config().Dependencies {
		if _, ok := builtModules[dep]; !ok {
			logger.Warnf("build skipped because dependency %q failed to build", dep)
			return nil
		}
	}

	meta.lastBuildStartTime = time.Now()
	e.projectMetas.Store(key, meta)
	err := e.build(ctx, key, builtModules, schemas)
	if err == nil && callback != nil {
		return callback(ctx, meta.project)
	}

	return err
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
	meta, ok := e.projectMetas.Load(key)
	if !ok {
		return fmt.Errorf("project %q not found", key)
	}

	combined := map[string]*schema.Module{}
	if err := e.gatherSchemas(builtModules, meta.project, combined); err != nil {
		return err
	}
	sch := &schema.Schema{Modules: maps.Values(combined)}

	if e.listener != nil {
		e.listener.OnBuildStarted(meta.project)
	}
	err := Build(ctx, sch, meta.project, e.watcher.GetTransaction(meta.project.Config().Dir))
	if err != nil {
		return err
	}
	if module, ok := meta.project.(Module); ok {
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
	latestModule, ok := e.projectMetas.Load(project.Config().Key)
	if !ok {
		latestModule = projectMeta{project: project}
	}
	for _, dep := range latestModule.project.Config().Dependencies {
		out[dep] = moduleSchemas[dep]
		if dep != "builtin" {
			depModule, ok := e.projectMetas.Load(ProjectKey(dep))
			// TODO: should we be gathering schemas from dependencies without a project?
			// This can happen if the schema is loaded from the controller
			if ok {
				if err := e.gatherSchemas(moduleSchemas, depModule.project, out); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
