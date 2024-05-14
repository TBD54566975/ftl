package configuration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
)

// InlineProvider is a configuration provider that stores configuration in its key.
type InlineProvider[R Role] struct {
	Inline bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"configwriter"`
}

var _ MutableProvider[Configuration] = InlineProvider[Configuration]{}

func (InlineProvider[R]) Role() R     { var r R; return r }
func (InlineProvider[R]) Key() string { return "inline" }

func (i InlineProvider[R]) Writer() bool { return i.Inline }

func (InlineProvider[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data in inline configuration: %w", err)
	}
	return data, nil
}

func (InlineProvider[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	b64 := base64.RawURLEncoding.EncodeToString(value)
	return &url.URL{Scheme: "inline", Host: b64}, nil
}

func (InlineProvider[R]) Delete(ctx context.Context, ref Ref) error {
	return nil
}
