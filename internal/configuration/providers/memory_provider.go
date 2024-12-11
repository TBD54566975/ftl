package providers

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/internal/configuration"
)

const MemoryProviderKey configuration.ProviderKey = "memory"

// Memory is a configuration provider that stores configuration in memory.
type Memory[R configuration.Role] struct {
	config map[string][]byte
}

var _ configuration.SynchronousProvider[configuration.Configuration] = &Memory[configuration.Configuration]{}

func NewMemory[R configuration.Role]() *Memory[R] {
	return &Memory[R]{config: map[string][]byte{}}
}

func NewMemoryFactory[R configuration.Role]() (configuration.ProviderKey, Factory[R]) {
	return MemoryProviderKey, func(ctx context.Context) (configuration.Provider[R], error) {
		return NewMemory[R](), nil
	}
}

func (m *Memory[R]) Role() R                        { var r R; return r }
func (m *Memory[R]) Key() configuration.ProviderKey { return MemoryProviderKey }

func (m *Memory[R]) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	return m.config[ref.String()], nil
}

func (m *Memory[R]) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	m.config[ref.String()] = value
	return &url.URL{Scheme: string(MemoryProviderKey), Host: ref.String()}, nil
}

func (m *Memory[R]) Delete(ctx context.Context, ref configuration.Ref) error {
	delete(m.config, ref.String())
	return nil
}
