package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/providers"
	"github.com/block/ftl/internal/configuration/routers"
	"github.com/block/ftl/internal/projectconfig"
)

// Manager is a high-level configuration manager that abstracts the details of
// the Router and Provider interfaces.
type Manager[R configuration.Role] struct {
	provider configuration.Provider[R]
	router   configuration.Router[R]
	cache    *cache[R]
}

func ConfigFromEnvironment() []string {
	if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
		return strings.Split(envar, ",")
	}
	return nil
}

// NewDefaultSecretsManagerFromConfig creates a new secrets manager from
// the project config found in the config paths.
func NewDefaultSecretsManagerFromConfig(ctx context.Context, registry *providers.Registry[configuration.Secrets], projectConfig projectconfig.Config) (*Manager[configuration.Secrets], error) {
	provider, err := registry.Get(ctx, projectConfig.SecretsProvider)
	if err != nil {
		return nil, fmt.Errorf("could not construct default secrets manager: %w", err)
	}
	var cr configuration.Router[configuration.Secrets] = routers.ProjectConfig[configuration.Secrets]{Config: projectConfig.Path}
	return New(ctx, cr, provider)
}

// NewDefaultConfigurationManagerFromConfig creates a new configuration manager from
// the project config found in the config paths.
func NewDefaultConfigurationManagerFromConfig(ctx context.Context, registry *providers.Registry[configuration.Configuration], projectConfig projectconfig.Config) (*Manager[configuration.Configuration], error) {
	provider, err := registry.Get(ctx, projectConfig.ConfigProvider)
	if err != nil {
		return nil, fmt.Errorf("could not construct default secrets manager: %w", err)
	}
	cr := routers.ProjectConfig[configuration.Configuration]{Config: projectConfig.Path}
	return New(ctx, cr, provider)
}

// New configuration manager.
func New[R configuration.Role](ctx context.Context, router configuration.Router[R], provider configuration.Provider[R]) (*Manager[R], error) {
	m := &Manager[R]{provider: provider}
	m.router = router

	var asyncProvider optional.Option[configuration.AsynchronousProvider[R]]
	if ap, ok := provider.(configuration.AsynchronousProvider[R]); ok {
		asyncProvider = optional.Some(ap)
	}
	m.cache = newCache[R](ctx, asyncProvider, m)
	return m, nil
}

func ProviderKeyForAccessor(accessor *url.URL) configuration.ProviderKey {
	return configuration.ProviderKey(accessor.Scheme)
}

// getData returns a data value for a configuration from the active providers.
// The data can be unmarshalled from JSON.
func (m *Manager[R]) getData(ctx context.Context, ref configuration.Ref) ([]byte, error) {
	key, err := m.router.Get(ctx, ref)
	// Try again at the global scope if the value is not found in module scope.
	if ref.Module.Ok() && errors.Is(err, configuration.ErrNotFound) {
		ref.Module = optional.None[string]()
		key, err = m.router.Get(ctx, ref)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if m.provider.Key() != ProviderKeyForAccessor(key) {
		return nil, fmt.Errorf("no provider for scheme %q", key.Scheme)
	}
	var data []byte
	switch provider := m.provider.(type) {
	case configuration.AsynchronousProvider[R]:
		data, err = m.cache.load(ref, key)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ref, err)
		}
	case configuration.SynchronousProvider[R]:
		data, err = provider.Load(ctx, ref, key)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ref, err)
		}
	default:
		if err != nil {
			return nil, fmt.Errorf("provider for %s does not support on demand access or syncing", ref)
		}
	}
	return data, nil
}

// Get a configuration value from the active providers.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m *Manager[R]) Get(ctx context.Context, ref configuration.Ref, value any) error {
	data, err := m.getData(ctx, ref)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}
	return nil
}

// ProviderKeys returns the keys of the available providers.
func (m *Manager[R]) ProviderKey() configuration.ProviderKey {
	return m.provider.Key()
}

// Set a configuration value, encoding "value" as JSON before storing it.
func (m *Manager[R]) Set(ctx context.Context, ref configuration.Ref, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.SetJSON(ctx, ref, data)
}

// SetJSON sets a configuration value using raw JSON data.
func (m *Manager[R]) SetJSON(ctx context.Context, ref configuration.Ref, value json.RawMessage) error {
	if err := checkJSON(value); err != nil {
		return fmt.Errorf("invalid value for %s, must be JSON: %w", m.router.Role(), err)
	}
	key, err := m.provider.Store(ctx, ref, value)
	if err != nil {
		return err
	}
	m.cache.updatedValue(ref, value, key)
	return m.router.Set(ctx, ref, key)
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
func (m *Manager[R]) Unset(ctx context.Context, ref configuration.Ref) error {
	if err := m.provider.Delete(ctx, ref); err != nil && !errors.Is(err, configuration.ErrNotFound) {
		return err
	}
	m.cache.deletedValue(ref, m.provider.Key())
	return m.router.Unset(ctx, ref)
}

func (m *Manager[R]) List(ctx context.Context) ([]configuration.Entry, error) {
	return m.router.List(ctx)
}

func checkJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var v any
	return dec.Decode(&v)
}
