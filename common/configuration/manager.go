package configuration

import (
	"bytes"
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
// the Router and Provider interfaces.
type Manager[R Role] struct {
	providers map[string]Provider[R]
	router    Router[R]
}

func ConfigFromEnvironment() []string {
	if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
		return strings.Split(envar, ",")
	}
	return nil
}

// NewDefaultSecretsManagerFromConfig creates a new secrets manager from
// the project config found in the config paths.
func NewDefaultSecretsManagerFromConfig(ctx context.Context, config string, opVault string) (*Manager[Secrets], error) {
	var cr Router[Secrets] = ProjectConfigResolver[Secrets]{Config: config}
	return NewSecretsManager(ctx, cr, opVault)
}

// NewDefaultConfigurationManagerFromConfig creates a new configuration manager from
// the project config found in the config paths.
func NewDefaultConfigurationManagerFromConfig(ctx context.Context, config string) (*Manager[Configuration], error) {
	cr := ProjectConfigResolver[Configuration]{Config: config}
	return NewConfigurationManager(ctx, cr)
}

// New configuration manager.
func New[R Role](ctx context.Context, router Router[R], providers []Provider[R]) (*Manager[R], error) {
	m := &Manager[R]{
		providers: map[string]Provider[R]{},
	}
	for _, p := range providers {
		m.providers[p.Key()] = p
	}
	m.router = router
	return m, nil
}

// getData returns a data value for a configuration from the active providers.
// The data can be unmarshalled from JSON.
func (m *Manager[R]) getData(ctx context.Context, ref Ref) ([]byte, error) {
	key, err := m.router.Get(ctx, ref)
	// Try again at the global scope if the value is not found in module scope.
	if ref.Module.Ok() && errors.Is(err, ErrNotFound) {
		ref.Module = optional.None[string]()
		key, err = m.router.Get(ctx, ref)
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
	err = json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("could not unmarshal: %w", err)
	}
	return nil
}

func (m *Manager[R]) availableProviderKeys() []string {
	keys := make([]string, 0, len(m.providers))
	for k := range m.providers {
		keys = append(keys, "--"+k)
	}
	return keys
}

// Set a configuration value, encoding "value" as JSON before storing it.
func (m *Manager[R]) Set(ctx context.Context, pkey string, ref Ref, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.SetJSON(ctx, pkey, ref, data)
}

// SetJSON sets a configuration value using raw JSON data.
func (m *Manager[R]) SetJSON(ctx context.Context, pkey string, ref Ref, value json.RawMessage) error {
	if err := checkJSON(value); err != nil {
		return fmt.Errorf("invalid value for %s, must be JSON: %w", m.router.Role(), err)
	}
	provider, err := m.selectProvider(pkey)
	if err != nil {
		return err
	}
	key, err := provider.Store(ctx, ref, value)
	if err != nil {
		return err
	}
	return m.router.Set(ctx, ref, key)
}

// selectProvider selects the Provider using the supplied key or defaults to the sole provider
// when no key is supplied and there exists exactly one provider.
func (m *Manager[R]) selectProvider(key string) (Provider[R], error) {
	if len(key) == 0 && len(m.availableProviderKeys()) == 1 {
		key = m.availableProviderKeys()[0]
	}
	provider, ok := m.providers[key]
	if !ok {
		keys := strings.Join(m.availableProviderKeys(), ", ")
		return nil, fmt.Errorf("no provider for key %q, specify one of: %s", key, keys)
	}
	return provider, nil
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
		pkeys := strings.Join(m.availableProviderKeys(), ", ")
		return fmt.Errorf("no provider for key %q, specify one of %s", pkey, pkeys)
	}
	if err := provider.Delete(ctx, ref); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return m.router.Unset(ctx, ref)
}

func (m *Manager[R]) List(ctx context.Context) ([]Entry, error) {
	return m.router.List(ctx)
}

func checkJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var v any
	return dec.Decode(&v)
}
