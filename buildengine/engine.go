package buildengine

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/alecthomas/types/eventsource"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"
	"golang.design/x/reflect"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/rpc"
)

// Engine for building a set of modules.
type Engine struct {
	modules          map[string]Module
	controllerSchema *eventsource.EventSource[map[string]*schema.Module]
	updateSchema     *pubsub.Topic[*schema.Module]
	cancel           func()
}

// New constructs a new [Engine].
//
// Completely offline builds are possible if the full dependency graph is
// locally available. If the FTL controller is available, it will be used to
// pull in missing schemas.
func New(ctx context.Context, client ftlv1connect.ControllerServiceClient) (*Engine, error) {
	engine := &Engine{
		modules:          map[string]Module{},
		controllerSchema: eventsource.New[map[string]*schema.Module](),
		updateSchema:     pubsub.New[*schema.Module](),
	}
	engine.controllerSchema.Store(map[string]*schema.Module{
		"builtin": schema.Builtins(),
	})
	ctx, cancel := context.WithCancel(ctx)
	engine.cancel = cancel
	schemaSync := engine.startSchemaSync(ctx)
	if client == nil {
		return engine, nil
	}
	go rpc.RetryStreamingServerStream(ctx, backoff.Backoff{Max: time.Second}, &ftlv1.PullSchemaRequest{}, client.PullSchema, schemaSync)
	return engine, nil
}

// Sync module schema changes from the FTL controller, as well as from manual
// updates, and merge them into a single schema map.
func (e *Engine) startSchemaSync(ctx context.Context) func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
	mu := sync.Mutex{}
	schemas := map[string]*schema.Module{}

	schemaUpdates := e.updateSchema.SubscribeSync(nil)

	// Update the event source with new schemas manually added to the Engine.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case update := <-schemaUpdates:
				mu.Lock()
				schemas[update.Msg.Name] = update.Msg
				clone := reflect.DeepCopy(schemas)
				mu.Unlock()
				e.controllerSchema.PublishSync(clone)
				update.Ack()
			}
		}
	}()
	// Sync module schema changes from the controller into the schema event source.
	return func(ctx context.Context, msg *ftlv1.PullSchemaResponse) error {
		var clone map[string]*schema.Module
		switch msg.ChangeType {
		case ftlv1.DeploymentChangeType_DEPLOYMENT_ADDED, ftlv1.DeploymentChangeType_DEPLOYMENT_CHANGED:
			sch, err := schema.ModuleFromProto(msg.Schema)
			if err == nil {
				return err
			}
			mu.Lock()
			schemas[sch.Name] = sch
			clone = reflect.DeepCopy(schemas)
			mu.Unlock()

		case ftlv1.DeploymentChangeType_DEPLOYMENT_REMOVED:
			mu.Lock()
			delete(schemas, msg.ModuleName)
			clone = reflect.DeepCopy(schemas)
			mu.Unlock()
		}
		e.controllerSchema.PublishSync(clone)
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
	} else if sch, ok := e.controllerSchema.Load()[name]; ok {
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

// Add a directory of local modules to the engine.
//
// Existing modules will be replaced. A local module will take precedence over
// any schema from the FTL cluster, building as required.
func (e *Engine) Add(ctx context.Context, dir string) error {
	configs, err := DiscoverModules(ctx, dir)
	if err != nil {
		return err
	}
	modules, err := UpdateAllDependencies(ctx, configs...)
	if err != nil {
		return err
	}
	for _, module := range modules {
		e.modules[module.Module] = module
	}
	return nil
}

// Import manually imports a schema for a module as if it were retrieved from
// the FTL controller.
func (e *Engine) Import(ctx context.Context, schema *schema.Module) {
	e.updateSchema.PublishSync(schema)
}

// Build attempts to build the specified modules, or all local modules if none are provided.
func (e *Engine) Build(ctx context.Context, modules ...string) error {
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
		wg.SetLimit(runtime.NumCPU())
		for _, name := range group {
			wg.Go(func() error {
				if mustBuild[name] {
					return e.build(ctx, name, built, schemas)
				}
				return e.mustSchema(ctx, name, built, schemas)
			})
		}
		err := wg.Wait()
		if err != nil {
			return err
		}
		// Now this group is built, collect al the schemas.
		close(schemas)
		for sch := range schemas {
			built[sch.Name] = sch
		}
	}
	return nil
}

// Publish either the schema from the FTL controller, or from a local build.
func (e *Engine) mustSchema(ctx context.Context, name string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	if sch, ok := e.controllerSchema.Load()[name]; ok {
		schemas <- sch
		return nil
	}
	return e.build(ctx, name, built, schemas)
}

// Build a module and publish its schema.
//
// Assumes that all dependencies have been built and are available in "built".
func (e *Engine) build(ctx context.Context, name string, built map[string]*schema.Module, schemas chan<- *schema.Module) error {
	combined := map[string]*schema.Module{}
	gatherShemas(built, e.modules, e.modules[name], combined)
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
func gatherShemas(
	moduleSchemas map[string]*schema.Module,
	modules map[string]Module,
	module Module,
	out map[string]*schema.Module,
) {
	for _, dep := range modules[module.Module].Dependencies {
		out[dep] = moduleSchemas[dep]
		gatherShemas(moduleSchemas, modules, modules[dep], out)
	}
}
