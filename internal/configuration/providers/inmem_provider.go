package providers

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

const InMemProviderKey configuration.ProviderKey = "inmem"

// InMem is a secret provider that stores values in memory.
//
// This should not be used in production.
type InMem[R configuration.Role] struct {
	secrets map[configuration.Ref][]byte
}

func NewInMem[R configuration.Role]() InMem[R] {
	return InMem[R]{
		secrets: make(map[configuration.Ref][]byte),
	}
}

func NewInMemFactory[R configuration.Role]() (configuration.ProviderKey, Factory[R]) {
	return InMemProviderKey, func(ctx context.Context) (configuration.Provider[R], error) {
		return NewInMem[R](), nil
	}
}

var _ configuration.SynchronousProvider[configuration.Secrets] = InMem[configuration.Secrets]{}

func (o InMem[R]) Role() R                        { return R{} }
func (o InMem[R]) Key() configuration.ProviderKey { return InMemProviderKey }
func (o InMem[R]) Delete(ctx context.Context, ref configuration.Ref) error {
	delete(o.secrets, ref)
	return nil
}

func (o InMem[R]) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	logger := log.FromContext(ctx)
	logger.Infof("loading secret: %s -> %s", ref, string(o.secrets[ref]))
	return o.secrets[ref], nil
}

func (o InMem[R]) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	logger := log.FromContext(ctx)
	logger.Infof("storing secret: %s -> %s", ref, string(value))

	o.secrets[ref] = value
	url := &url.URL{Scheme: string(InMemProviderKey)}

	return url, nil
}
