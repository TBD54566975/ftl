package configuration

import (
	"context"
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
	return []byte(key.Host), nil
}

func (DBProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	return &url.URL{Scheme: "db", Host: string(value)}, nil
}

func (DBProvider) Delete(ctx context.Context, ref Ref) error {
	return nil
}
