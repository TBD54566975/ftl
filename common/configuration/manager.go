package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	writer    MutableProvider[R]
	resolver  Resolver[R]
}

// New configuration manager.
func New[R Role](ctx context.Context, resolver Resolver[R], providers []Provider[R]) (*Manager[R], error) {
	m := &Manager[R]{
		providers: map[string]Provider[R]{},
	}
	for _, p := range providers {
		m.providers[p.Key()] = p
		if mutable, ok := p.(MutableProvider[R]); ok && mutable.Writer() {
			if m.writer != nil {
				return nil, fmt.Errorf("multiple writers %s and %s", m.writer.Key(), p.Key())
			}
			m.writer = mutable
		}
	}
	m.resolver = resolver
	return m, nil
}

// Mutable returns an error if the configuration manager doesn't have a
// writeable provider configured.
func (m *Manager[R]) Mutable() error {
	if m.writer != nil {
		return nil
	}
	writers := []string{}
	for _, p := range m.providers {
		if mutable, ok := p.(MutableProvider[R]); ok {
			writers = append(writers, "--"+mutable.Key())
		}
	}
	return fmt.Errorf("no writeable configuration provider available, specify one of %s", strings.Join(writers, ", "))
}

// GetData returns a data value for a configuration from the active providers.
// The data can be unmarshalled from JSON.
func (m *Manager[R]) GetData(ctx context.Context, ref Ref) ([]byte, error) {
	key, err := m.resolver.Get(ctx, ref)
	// Try again at the global scope if the value is not found in module scope.
	if ref.Module.Ok() && errors.Is(err, ErrNotFound) {
		gref := ref
		gref.Module = optional.None[string]()
		key, err = m.resolver.Get(ctx, gref)
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
	data, err := m.GetData(ctx, ref)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// SetData updates the configuration value in the active writing provider as raw bytes.
//
// "data" must be bytes that can be unmarshalled into a Go type.
func (m *Manager[R]) SetData(ctx context.Context, ref Ref, data []byte) error {
	if err := m.Mutable(); err != nil {
		return err
	}
	key, err := m.writer.Store(ctx, ref, data)
	if err != nil {
		return err
	}
	return m.resolver.Set(ctx, ref, key)
}

// Set a configuration value in the active writing provider.
//
// "value" must be a Go type that can be marshalled to JSON.
func (m *Manager[R]) Set(ctx context.Context, ref Ref, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.SetData(ctx, ref, data)
}

// Unset a configuration value in all providers.
func (m *Manager[R]) Unset(ctx context.Context, ref Ref) error {
	for _, provider := range m.providers {
		if mutable, ok := provider.(MutableProvider[R]); ok {
			if err := mutable.Delete(ctx, ref); err != nil && !errors.Is(err, ErrNotFound) {
				return err
			}
		}
	}
	return m.resolver.Unset(ctx, ref)
}

func (m *Manager[R]) List(ctx context.Context) ([]Entry, error) {
	return m.resolver.List(ctx)
}
