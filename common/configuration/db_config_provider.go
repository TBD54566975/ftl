package configuration

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/db/dalerrs"
	"github.com/alecthomas/types/optional"
)

// DBConfigProvider is a configuration provider that stores configuration in its key.
type DBConfigProvider struct {
	dal DBConfigProviderDAL
}

type DBConfigProviderDAL interface {
	GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error
}

func NewDBConfigProvider(dal DBConfigProviderDAL) DBConfigProvider {
	return DBConfigProvider{
		dal: dal,
	}
}

func (DBConfigProvider) Role() Configuration { return Configuration{} }
func (DBConfigProvider) Key() string         { return "db" }

func (d DBConfigProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	value, err := d.dal.GetModuleConfiguration(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, dalerrs.ErrNotFound
	}
	return value, nil
}

func (d DBConfigProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	err := d.dal.SetModuleConfiguration(ctx, ref.Module, ref.Name, value)
	if err != nil {
		return nil, err
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DBConfigProvider) Delete(ctx context.Context, ref Ref) error {
	return d.dal.UnsetModuleConfiguration(ctx, ref.Module, ref.Name)
}
