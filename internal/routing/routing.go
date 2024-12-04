package routing

import (
	"context"
	"net/url"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

type RouteTable struct {
	// Routes keyed by module name. TODO: this should be keyed by deployment key.
	routes *atomic.Value[map[string]*url.URL]
}

func New(ctx context.Context, changes schemaeventsource.EventSource) *RouteTable {
	r := &RouteTable{
		routes: atomic.New(extractRoutes(ctx, changes.View())),
	}
	go r.run(ctx, changes)
	return r
}

func (r *RouteTable) run(ctx context.Context, changes schemaeventsource.EventSource) {
	for {
		select {
		case <-ctx.Done():
			return

		case event := <-changes.Events():
			routes := extractRoutes(ctx, event.Schema())
			r.routes.Store(routes)
		}
	}
}

// Get returns the URL for the given deployment or None if it doesn't exist.
func (r *RouteTable) Get(deployment model.DeploymentKey) optional.Option[*url.URL] {
	routes := r.routes.Load()
	return optional.Zero(routes[deployment.Payload.Module])
}

// GetForModule returns the URL for the given module or None if it doesn't exist.
func (r *RouteTable) GetForModule(module string) optional.Option[*url.URL] {
	routes := r.routes.Load()
	return optional.Zero(routes[module])
}

func extractRoutes(ctx context.Context, schema *schema.Schema) map[string]*url.URL {
	logger := log.FromContext(ctx)
	out := make(map[string]*url.URL, len(schema.Modules))
	for _, module := range schema.Modules {
		if module.Runtime == nil || module.Runtime.Deployment == nil {
			continue
		}
		rt := module.Runtime.Deployment
		u, err := url.Parse(rt.Endpoint)
		if err != nil {
			logger.Warnf("Failed to parse endpoint URL for module %q: %v", module.Name, err)
			continue
		}
		out[module.Name] = u
	}
	return out
}
