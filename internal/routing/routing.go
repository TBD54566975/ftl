package routing

import (
	"context"
	"net/url"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type RouteView struct {
	byDeployment       map[string]*url.URL
	moduleToDeployment map[string]model.DeploymentKey
	schema             *schema.Schema
}

type RouteTable struct {
	routes *atomic.Value[RouteView]
	// When the routes for a module change they are published here.
	changeNotification *pubsub.Topic[string]
}

func New(ctx context.Context, changes schemaeventsource.EventSource) *RouteTable {
	r := &RouteTable{
		routes:             atomic.New(extractRoutes(ctx, changes.View())),
		changeNotification: pubsub.New[string](),
	}
	go r.run(ctx, changes)
	return r
}

func (r *RouteTable) run(ctx context.Context, changes schemaeventsource.EventSource) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-changes.Events():
			old := r.routes.Load()
			routes := extractRoutes(ctx, changes.View())
			for module, rd := range old.moduleToDeployment {
				if old.byDeployment[rd.String()] != routes.byDeployment[rd.String()] {
					r.changeNotification.Publish(module)
				}
			}
			for module, rd := range routes.moduleToDeployment {
				// Check for new modules
				if old.byDeployment[rd.String()] == nil {
					r.changeNotification.Publish(module)
				}
			}
			r.routes.Store(routes)
		}
	}
}

// Current returns the current routes.
func (r *RouteTable) Current() RouteView {
	return r.routes.Load()
}

// Get returns the URL for the given deployment or None if it doesn't exist.
func (r RouteView) Get(deployment model.DeploymentKey) optional.Option[url.URL] {
	mod := r.byDeployment[deployment.String()]
	if mod == nil {
		return optional.None[url.URL]()
	}
	return optional.Some(*mod)
}

// GetForModule returns the URL for the given module or None if it doesn't exist.
func (r RouteView) GetForModule(module string) optional.Option[url.URL] {
	dep, ok := r.moduleToDeployment[module]
	if !ok {
		return optional.None[url.URL]()
	}
	return r.Get(dep)
}

// GetDeployment returns the deployment key for the given module or None if it doesn't exist.
func (r RouteView) GetDeployment(module string) optional.Option[model.DeploymentKey] {
	return optional.Zero(r.moduleToDeployment[module])
}

// Schema returns the current schema that the routes are based on.
func (r RouteView) Schema() *schema.Schema {
	return r.schema
}

func (r *RouteTable) Subscribe() chan string {
	return r.changeNotification.Subscribe(nil)
}
func (r *RouteTable) Unsubscribe(s chan string) {
	r.changeNotification.Unsubscribe(s)
}

func extractRoutes(ctx context.Context, sch *schema.Schema) RouteView {
	if sch == nil {
		return RouteView{moduleToDeployment: map[string]model.DeploymentKey{}, byDeployment: map[string]*url.URL{}, schema: &schema.Schema{}}
	}
	logger := log.FromContext(ctx)
	moduleToDeployment := make(map[string]model.DeploymentKey, len(sch.Modules))
	byDeployment := make(map[string]*url.URL, len(sch.Modules))
	for _, module := range sch.Modules {
		if module.Runtime == nil || module.Runtime.Deployment == nil {
			continue
		}
		rt := module.Runtime.Deployment
		key, err := model.ParseDeploymentKey(rt.DeploymentKey)
		if err != nil {
			logger.Warnf("Failed to parse deployment key for module %q: %v", module.Name, err)
			continue
		}
		u, err := url.Parse(rt.Endpoint)
		if err != nil {
			logger.Warnf("Failed to parse endpoint URL for module %q: %v", module.Name, err)
			continue
		}
		logger.Debugf("Adding route for %s/%s: %s", module.Name, rt.DeploymentKey, u)
		moduleToDeployment[module.Name] = key
		byDeployment[rt.DeploymentKey] = u
	}
	return RouteView{moduleToDeployment: moduleToDeployment, byDeployment: byDeployment, schema: sch}
}
