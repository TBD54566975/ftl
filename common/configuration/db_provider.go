package configuration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
)

// DBProvider is a configuration provider that stores configuration in its key.
type DBProvider struct {
	DB bool `help:"Write configuration values to the database." group:"Provider:" xor:"configwriter"`
}

var _ MutableProvider[Configuration] = DBProvider{}

func (DBProvider) Role() Configuration { return Configuration{} }
func (DBProvider) Key() string         { return "db" }

func (d DBProvider) Writer() bool { return d.DB }

func (DBProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 data in db configuration: %w", err)
	}
	return data, nil
}

func (DBProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	b64 := base64.RawURLEncoding.EncodeToString(value)
	return &url.URL{Scheme: "db", Host: b64}, nil
}

func (DBProvider) Delete(ctx context.Context, ref Ref) error {
	return nil
}
