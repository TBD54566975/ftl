package configuration

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/alecthomas/types/optional"
)

// DBConfigProvider is a configuration provider that stores configuration in its key.
type DBConfigProvider struct {
	DB  bool `help:"Write configuration values to the database." group:"Provider:" xor:"configwriter"`
	dal DBConfigProviderDAL
}

type DBConfigProviderDAL interface {
	GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error
}

var _ MutableProvider[Configuration] = DBConfigProvider{}

func NewDBConfigProvider(dal DBConfigProviderDAL) DBConfigProvider {
	return DBConfigProvider{
		DB:  true,
		dal: dal,
	}
}

func (DBConfigProvider) Role() Configuration { return Configuration{} }
func (DBConfigProvider) Key() string         { return "db" }

func (d DBConfigProvider) Writer() bool { return d.DB }

func (d DBConfigProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	value, err := d.dal.GetModuleConfiguration(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, dal.ErrNotFound
	}
	return value, nil
}

func (d DBConfigProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	err := d.dal.SetModuleConfiguration(ctx, ref.Module, ref.Name, value)
	if err != nil {
		return nil, dal.TranslatePGError(err)
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DBConfigProvider) Delete(ctx context.Context, ref Ref) error {
	err := d.dal.UnsetModuleConfiguration(ctx, ref.Module, ref.Name)
	return dal.TranslatePGError(err)
}
