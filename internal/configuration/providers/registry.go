package providers

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/TBD54566975/ftl/internal/configuration"
)

type Factory[R configuration.Role] func(ctx context.Context) (configuration.Provider[R], error)

// Registry that lazily constructs configuration providers.
type Registry[R configuration.Role] struct {
	factories map[configuration.ProviderKey]Factory[R]
}

func NewRegistry[R configuration.Role]() *Registry[R] {
	return &Registry[R]{
		factories: map[configuration.ProviderKey]Factory[R]{},
	}
}

// Providers returns the list of registered provider keys.
func (r *Registry[R]) Providers() []configuration.ProviderKey {
	return slices.Collect(maps.Keys(r.factories))
}

func (r *Registry[R]) Register(name configuration.ProviderKey, factory Factory[R]) {
	r.factories[name] = factory
}

func (r *Registry[R]) Get(ctx context.Context, name configuration.ProviderKey) (configuration.Provider[R], error) {
	factory, ok := r.factories[name]
	if !ok {
		var role R
		return nil, fmt.Errorf("%s: %s provider not found", name, role)
	}
	return factory(ctx)
}
