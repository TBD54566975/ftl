package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/block/ftl/internal/configuration"
)

const InlineProviderKey configuration.ProviderKey = "inline"

// Inline is a configuration provider that stores configuration in its key.
type Inline[R configuration.Role] struct{}

var _ configuration.SynchronousProvider[configuration.Configuration] = Inline[configuration.Configuration]{}

func NewInline[R configuration.Role]() Inline[R] {
	return Inline[R]{}
}

func NewInlineFactory[R configuration.Role]() (configuration.ProviderKey, Factory[R]) {
	return InlineProviderKey, func(ctx context.Context) (configuration.Provider[R], error) {
		return NewInline[R](), nil
	}
}

func (Inline[R]) Role() R                        { var r R; return r }
func (Inline[R]) Key() configuration.ProviderKey { return InlineProviderKey }

func (Inline[R]) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data in inline configuration: %w", err)
	}
	return data, nil
}

func (Inline[R]) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	b64 := base64.RawURLEncoding.EncodeToString(value)
	return &url.URL{Scheme: string(InlineProviderKey), Host: b64}, nil
}

func (Inline[R]) Delete(ctx context.Context, ref configuration.Ref) error {
	return nil
}
