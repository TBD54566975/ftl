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
	modules          map[string]Module
	dirs             []string
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
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient, dirs []string, options ...Option) (*Engine, error) {
	ctx = rpc.ContextWithClient(ctx, client)
	e := &Engine{
		client:           client,
		dirs:             dirs,
		modules:          map[string]Module{},
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
	configs, err := DiscoverModules(ctx, dirs...)
	if err != nil {
		return nil, err
	}
	modules, err := UpdateAllDependencies(ctx, configs...)
	if err != nil {
		return nil, err
	}
	for _, module := range modules {
		e.modules[module.Module] = module
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
func (e *Engine) Graph(modules ...string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(modules) == 0 {
		modules = maps.Keys(e.modules)
	}
	for _, name := range modules {
		if err := e.buildGraph(name, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Engine) buildGraph(name string, out map[string][]string) error {
	var deps []string
	if module, ok := e.modules[name]; ok {
		deps = module.Dependencies
	} else if sch, ok := e.controllerSchema.Load(name); ok {
		deps = sch.Imports()
	} else {
		return fmt.Errorf("module %q not found", name)
	}
	out[name] = deps
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
	watch := Watch(ctx, period, e.dirs...)
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
			case WatchEventModuleAdded:
				if _, exists := e.modules[event.Module.Module]; !exists {
					e.modules[event.Module.Module] = event.Module
					err := e.buildAndDeploy(ctx, 1, true, event.Module.Module)
					if err != nil {
						logger.Errorf(err, "deploy %s failed", event.Module.Module)
					}
				}
			case WatchEventModuleRemoved:
				err := teminateModuleDeployment(ctx, e.client, event.Module.Module)
				if err != nil {
					logger.Errorf(err, "terminate %s failed", event.Module.Module)
				}
				delete(e.modules, event.Module.Module)

			case WatchEventModuleChanged:
				err := e.buildAndDeploy(ctx, 1, true, event.Module.Module)
				if err != nil {
					logger.Errorf(err, "deploy %s failed", event.Module.Module)
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
			modulesToDeploy := e.getDependentModules(change.Name)
			if len(modulesToDeploy) > 0 {
				logger.Infof("%s's schema changed; redeploying %s", change.Name, strings.Join(modulesToDeploy, ", "))
				err = e.buildAndDeploy(ctx, 1, true, modulesToDeploy...)
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

func (e *Engine) getDependentModules(moduleName string) []string {
	dependentModules := map[string]bool{}
	for key, module := range e.modules {
		for _, dep := range module.Dependencies {
			if dep == moduleName {
				dependentModules[key] = true
			}
		}
	}

	return maps.Keys(dependentModules)
}

func (e *Engine) buildAndDeploy(ctx context.Context, replicas int32, waitForDeployOnline bool, modules ...string) error {
	if len(modules) == 0 {
		modules = maps.Keys(e.modules)
	}

	deployQueue := make(chan Module, len(modules))
	wg, ctx := errgroup.WithContext(ctx)

	// Build all modules and enqueue them for deployment.
	wg.Go(func() error {
		defer close(deployQueue)

		return e.buildWithCallback(ctx, func(ctx context.Context, module Module) error {
			select {
			case deployQueue <- module:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}, modules...)
	})

	// Process deployment queue.
	for i := 0; i < len(modules); i++ {
		wg.Go(func() error {
			for {
				select {
				case module, ok := <-deployQueue:
					if !ok {
						return nil
					}
					if err := Deploy(ctx, module, replicas, waitForDeployOnline, e.client); err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	return wg.Wait()
}

type buildCallback func(ctx context.Context, module Module) error

func (e *Engine) buildWithCallback(ctx context.Context, callback buildCallback, modules ...string) error {
	mustBuild := map[string]bool{}
	if len(modules) == 0 {
		modules = maps.Keys(e.modules)
	}
	for _, name := range modules {
		module, ok := e.modules[name]
		if !ok {
			return fmt.Errorf("module %q not found", module)
		}
		// Update dependencies before building.
		var err error
		module, err = UpdateDependencies(ctx, module.ModuleConfig)
		if err != nil {
			return err
		}
		e.modules[name] = module
		mustBuild[name] = true
	}
	graph, err := e.Graph(modules...)
	if err != nil {
		return err
	}
	built := map[string]*schema.Module{
		"builtin": schema.Builtins(),
	}

	topology := TopologicalSort(graph)
	for _, group := range topology {
		// Collect schemas to be inserted into "built" map for subsequent groups.
		schemas := make(chan *schema.Module, len(group))
		wg, ctx := errgroup.WithContext(ctx)
		wg.SetLimit(e.parallelism)
		for _, name := range group {
			wg.Go(func() error {
				if mustBuild[name] {
					err := e.build(ctx, name, built, schemas)
					if err == nil && callback != nil {
						return callback(ctx, e.modules[name])
					}
					return err
				}
				return e.mustSchema(ctx, name, built, schemas)
			})
		}
		err := wg.Wait()
		if err != nil {
			return err
		}
		// Now this group is built, collect all the schemas.
		close(schemas)
		for sch := range schemas {
			built[sch.Name] = sch
		}
	}
	return nil
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, name string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load(name); ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, name, built, schemas)
}

func (e *Engine) LoadSchemaFromController(name string) (*schema.Module, error) {

	if sch, ok := e.controllerSchema.Load(name); ok {
		return sch, nil
	}
	return nil, fmt.Errorf("schema for %s not found", name)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, name string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	combined := map[string]*schema.Module{}
	gatherSchemas(built, e.modules, e.modules[name], combined)
	sch := &schema.Schema{Modules: maps.Values(combined)}
	module := e.modules[name]
	err := Build(ctx, sch, module)
	if err != nil {

		return err
	}
	moduleSchema, err := schema.ModuleFromProtoFile(filepath.Join(module.Dir, module.DeployDir, module.Schema))
	if err != nil {
		return fmt.Errorf("load schema %s: %w", name, err)
	}
	schemas <- moduleSchema
	return nil
}

// Construct a combined schema for a module and its transitive dependencies.
func gatherSchemas(
	moduleSchemas map[string]*schema.Module,
	modules map[string]Module,
	module Module,
	out map[string]*schema.Module,
) {
	for _, dep := range modules[module.Module].Dependencies {
		out[dep] = moduleSchemas[dep]
		gatherSchemas(moduleSchemas, modules, modules[dep], out)
	}
}
