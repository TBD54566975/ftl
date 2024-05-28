package dal

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/common/configuration"
)

// DBConfigProvider is a configuration provider that stores configuration in its key.
type DBConfigProvider struct {
	DB    bool `help:"Write configuration values to the database." group:"Provider:" xor:"configwriter"`
	dalDB sql.DBI
}

var _ configuration.MutableProvider[configuration.Configuration] = DBConfigProvider{}

func (d *DAL) NewConfigProvider() DBConfigProvider {
	return DBConfigProvider{
		DB:    true,
		dalDB: d.db,
	}
}

func (DBConfigProvider) Role() configuration.Configuration { return configuration.Configuration{} }
func (DBConfigProvider) Key() string                       { return "db" }

func (d DBConfigProvider) Writer() bool { return d.DB }

func (d DBConfigProvider) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	value, err := d.dalDB.GetConfig(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, configuration.ErrNotFound
	}
	return value, nil
}

func (d DBConfigProvider) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	err := d.dalDB.SetConfig(ctx, ref.Module, ref.Name, value)
	if err != nil {
		return nil, translatePGError(err)
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DBConfigProvider) Delete(ctx context.Context, ref configuration.Ref) error {
	err := d.dalDB.UnsetConfig(ctx, ref.Module, ref.Name)
	return translatePGError(err)
}
