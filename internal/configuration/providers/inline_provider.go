package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/configuration"
)

// Inline is a configuration provider that stores configuration in its key.
type Inline[R configuration.Role] struct{}

var _ configuration.SynchronousProvider[configuration.Configuration] = Inline[configuration.Configuration]{}

func (Inline[R]) Role() R     { var r R; return r }
func (Inline[R]) Key() string { return "inline" }

func (Inline[R]) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data in inline configuration: %w", err)
	}
	return data, nil
}

func (Inline[R]) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	b64 := base64.RawURLEncoding.EncodeToString(value)
	return &url.URL{Scheme: "inline", Host: b64}, nil
}

func (Inline[R]) Delete(ctx context.Context, ref configuration.Ref) error {
	return nil
}
