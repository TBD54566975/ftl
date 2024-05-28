package configuration

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/dalerrors"
	"github.com/TBD54566975/ftl/backend/controller/sql"
)

// DBConfigProvider is a configuration provider that stores configuration in its key.
type DBConfigProvider struct {
	DB    bool `help:"Write configuration values to the database." group:"Provider:" xor:"configwriter"`
	dalDB sql.DBI
}

var _ MutableProvider[Configuration] = DBConfigProvider{}

func NewDBConfigProvider(db sql.DBI) DBConfigProvider {
	return DBConfigProvider{
		DB:    true,
		dalDB: db,
	}
}

func (DBConfigProvider) Role() Configuration { return Configuration{} }
func (DBConfigProvider) Key() string         { return "db" }

func (d DBConfigProvider) Writer() bool { return d.DB }

func (d DBConfigProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	value, err := d.dalDB.GetModuleConfiguration(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, dalerrors.ErrNotFound
	}
	return value, nil
}

func (d DBConfigProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	err := d.dalDB.SetModuleConfiguration(ctx, ref.Module, ref.Name, value)
	if err != nil {
		return nil, dalerrors.TranslatePGError(err)
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DBConfigProvider) Delete(ctx context.Context, ref Ref) error {
	err := d.dalDB.UnsetModuleConfiguration(ctx, ref.Module, ref.Name)
	return dalerrors.TranslatePGError(err)
}
