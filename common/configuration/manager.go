package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/types/optional"
)

// Role of [Manager], either Secrets or Configuration.
type Role interface {
	Secrets | Configuration
}

type Secrets struct{}

func (Secrets) String() string { return "secrets" }

type Configuration struct{}

func (Configuration) String() string { return "configuration" }

// Manager is a high-level configuration manager that abstracts the details of
// the Resolver and Provider interfaces.
type Manager[R Role] struct {
	providers map[string]Provider[R]
	resolver  Resolver[R]
}

func ConfigFromEnvironment() []string {
	if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
		return strings.Split(envar, ",")
	}
	return nil
}

// NewDefaultSecretsManagerFromConfig creates a new secrets manager from
// the project config found in the config paths.
func NewDefaultSecretsManagerFromConfig(ctx context.Context, config []string, opVault string) (*Manager[Secrets], error) {
	var cr Resolver[Secrets] = ProjectConfigResolver[Secrets]{Config: config}
	return NewSecretsManager(ctx, cr, opVault)
}

// NewDefaultConfigurationManagerFromConfig creates a new configuration manager from
// the project config found in the config paths.
func NewDefaultConfigurationManagerFromConfig(ctx context.Context, config []string) (*Manager[Configuration], error) {
	cr := ProjectConfigResolver[Configuration]{Config: config}
	return NewConfigurationManager(ctx, cr)
}

// New configuration manager.
func New[R Role](ctx context.Context, resolver Resolver[R], providers []Provider[R]) (*Manager[R], error) {
	m := &Manager[R]{
		providers: map[string]Provider[R]{},
	}
	for _, p := range providers {
		m.providers[p.Key()] = p
	}
	m.resolver = resolver
	return m, nil
}

// getData returns a data value for a configuration from the active providers.
// The data can be unmarshalled from JSON.
func (m *Manager[R]) getData(ctx context.Context, ref Ref) ([]byte, error) {
	key, err := m.resolver.Get(ctx, ref)
	// Try again at the global scope if the value is not found in module scope.
	if ref.Module.Ok() && errors.Is(err, ErrNotFound) {
		ref.Module = optional.None[string]()
		key, err = m.resolver.Get(ctx, ref)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	provider, ok := m.providers[key.Scheme]
	if !ok {
		return nil, fmt.Errorf("no provider for scheme %q", key.Scheme)
	}
	data, err := provider.Load(ctx, ref, key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ref, err)
	}
	return data, nil
}

// Get a configuration value from the active providers.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m *Manager[R]) Get(ctx context.Context, ref Ref, value any) error {
	data, err := m.getData(ctx, ref)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// Set a configuration value.
func (m *Manager[R]) Set(ctx context.Context, pkey string, ref Ref, value any) error {
	provider, ok := m.providers[pkey]
	if !ok {
		return fmt.Errorf("no provider for key %q", pkey)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	key, err := provider.Store(ctx, ref, data)
	if err != nil {
		return err
	}
	return m.resolver.Set(ctx, ref, key)
}

// MapForModule combines all configuration values visible to the module. Local
// values take precedence.
func (m *Manager[R]) MapForModule(ctx context.Context, module string) (map[string][]byte, error) {
	entries, err := m.List(ctx)
	if err != nil {
		return nil, err
	}
	combined := map[string][]byte{}
	locals := map[string][]byte{}
	for _, entry := range entries {
		mod, ok := entry.Module.Get()
		if ok && mod == module {
			data, err := m.getData(ctx, entry.Ref)
			if err != nil {
				return nil, err
			}
			locals[entry.Ref.Name] = data
		} else if !ok {
			data, err := m.getData(ctx, entry.Ref)
			if err != nil {
				return nil, err
			}
			combined[entry.Ref.Name] = data
		}
	}
	for k, v := range locals {
		combined[k] = v
	}
	return combined, nil
}

// Unset a configuration value in all providers.
func (m *Manager[R]) Unset(ctx context.Context, pkey string, ref Ref) error {
	provider, ok := m.providers[pkey]
	if !ok {
		return fmt.Errorf("no provider for key %q", pkey)
	}
	if err := provider.Delete(ctx, ref); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return m.resolver.Unset(ctx, ref)
}

func (m *Manager[R]) List(ctx context.Context) ([]Entry, error) {
	return m.resolver.List(ctx)
}
