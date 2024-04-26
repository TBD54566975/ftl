package configuration

import (
	"context"
	"net/url"
)

// InMemoryProvider is a configuration provider that keeps values in memory
type InMemoryProvider[R Role] struct {
	values map[Ref][]byte
}

var _ MutableProvider[Configuration] = &InMemoryProvider[Configuration]{}

func NewInMemoryProvider[R Role]() *InMemoryProvider[R] {
	return &InMemoryProvider[R]{values: map[Ref][]byte{}}
}

func (p *InMemoryProvider[R]) Role() R     { var r R; return r }
func (p *InMemoryProvider[R]) Key() string { return "inmemory" }

func (p *InMemoryProvider[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	if bytes, found := p.values[ref]; found {
		return bytes, nil
	}
	return nil, ErrNotFound
}

func (p *InMemoryProvider[R]) Writer() bool {
	return true
}

// Store a configuration value and return its key.
func (p *InMemoryProvider[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	p.values[ref] = value
	return &url.URL{Scheme: p.Key()}, nil
}

// Delete a configuration value.
func (p *InMemoryProvider[R]) Delete(ctx context.Context, ref Ref) error {
	delete(p.values, ref)
	return nil
}
