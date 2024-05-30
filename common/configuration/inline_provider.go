package configuration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/alecthomas/types/optional"
)

// InlineProvider is a configuration provider that stores configuration in its key.
type InlineProvider[R Role] struct{}

func (InlineProvider[R]) Role() R     { var r R; return r }
func (InlineProvider[R]) Key() string { return "inline" }

func (InlineProvider[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data in inline configuration: %w", err)
	}
	return data, nil
}

func (InlineProvider[R]) Store(ctx context.Context, host optional.Option[string], ref Ref, value []byte) (*url.URL, error) {
	if h, ok := host.Get(); ok && h != "" {
		return nil, fmt.Errorf("inline configuration does not support host: %s", h)
	}
	b64 := base64.RawURLEncoding.EncodeToString(value)
	return &url.URL{Scheme: "inline", Host: b64}, nil
}

func (InlineProvider[R]) Delete(ctx context.Context, ref Ref) error {
	return nil
}
