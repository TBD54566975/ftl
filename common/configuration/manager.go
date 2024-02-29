package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Manager is a high-level configuration manager that abstracts the details of
// the Resolver and Provider interfaces.
type Manager struct {
	providers map[string]Provider
	writer    MutableProvider
	resolver  Resolver
}

// New configuration manager.
func New(ctx context.Context, resolver Resolver, providers []Provider) (*Manager, error) {
	m := &Manager{
		providers: map[string]Provider{},
	}
	for _, p := range providers {
		m.providers[p.Key()] = p
		if mutable, ok := p.(MutableProvider); ok && mutable.Writer() {
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
func (m *Manager) Mutable() error {
	if m.writer != nil {
		return nil
	}
	writers := []string{}
	for _, p := range m.providers {
		if mutable, ok := p.(MutableProvider); ok {
			writers = append(writers, "--"+mutable.Key())
		}
	}
	return fmt.Errorf("no writeable configuration provider available, specify one of %s", strings.Join(writers, ", "))
}

// Get a configuration value from the active providers.
//
// "value" must be a pointer to a Go type that can be unmarshalled from JSON.
func (m *Manager) Get(ctx context.Context, ref Ref, value any) error {
	key, err := m.resolver.Get(ctx, ref)
	if err != nil {
		return err
	}
	provider, ok := m.providers[key.Scheme]
	if !ok {
		return fmt.Errorf("no provider for scheme %q", key.Scheme)
	}
	data, err := provider.Load(ctx, ref, key)
	if err != nil {
		return fmt.Errorf("%s: %w", ref, err)
	}
	return json.Unmarshal(data, value)
}

// Set a configuration value in the active writing provider.
//
// "value" must be a Go type that can be marshalled to JSON.
func (m *Manager) Set(ctx context.Context, ref Ref, value any) error {
	if err := m.Mutable(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	key, err := m.writer.Store(ctx, ref, data)
	if err != nil {
		return err
	}
	return m.resolver.Set(ctx, ref, key)
}

// Unset a configuration value in all providers.
func (m *Manager) Unset(ctx context.Context, ref Ref) error {
	for _, provider := range m.providers {
		if mutable, ok := provider.(MutableProvider); ok {
			if err := mutable.Delete(ctx, ref); err != nil && !errors.Is(err, ErrNotFound) {
				return err
			}
		}
	}
	return m.resolver.Unset(ctx, ref)
}

func (m *Manager) List(ctx context.Context) ([]Entry, error) {
	entries := []Entry{}
	for _, provider := range m.providers {
		if resolver, ok := provider.(Resolver); ok {
			subentries, err := resolver.List(ctx)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", provider.Key(), err)
			}
			entries = append(entries, subentries...)
		}
	}
	subentries, err := m.resolver.List(ctx)
	if err != nil {
		return nil, err
	}
	entries = append(entries, subentries...)
	return entries, nil
}
