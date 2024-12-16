package providers

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/projectconfig"
)

type Factory[R configuration.Role] func(ctx context.Context) (configuration.Provider[R], error)

// NewDefaultConfigRegistry creates a new registry with the default configuration providers.
func NewDefaultConfigRegistry() *Registry[configuration.Configuration] {
	registry := NewRegistry[configuration.Configuration]()
	registry.Register(NewEnvarFactory[configuration.Configuration]())
	registry.Register(NewInlineFactory[configuration.Configuration]())
	return registry
}

// NewDefaultSecretsRegistry creates a new registry with the default secrets providers.
func NewDefaultSecretsRegistry(config projectconfig.Config, onePasswordVault string) *Registry[configuration.Secrets] {
	registry := NewRegistry[configuration.Secrets]()
	registry.Register(NewEnvarFactory[configuration.Secrets]())
	registry.Register(NewInlineFactory[configuration.Secrets]())
	registry.Register(NewOnePasswordFactory(onePasswordVault, config.Name))
	registry.Register(NewKeychainFactory())
	return registry
}

// Registry that lazily constructs configuration
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
