package dal

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/slices"
)

// DatabaseResolver loads values a project's configuration from the given database.
type DatabaseResolver struct {
	db sql.DBI
}

// DatabaseResolver should only be used for config, not secrets
var _ configuration.Resolver[configuration.Configuration] = DatabaseResolver{}

func (d DatabaseResolver) Role() configuration.Configuration { return configuration.Configuration{} }

func (d *DAL) NewConfigResolver() DatabaseResolver {
	return DatabaseResolver{
		db: d.db,
	}
}

func (d DatabaseResolver) Get(ctx context.Context, ref configuration.Ref) (*url.URL, error) {
	value, err := d.db.GetConfig(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, configuration.ErrNotFound
	}
	p := configuration.DBProvider{true}
	return p.Store(ctx, ref, []byte(value))
}

func (d DatabaseResolver) List(ctx context.Context) ([]configuration.Entry, error) {
	configs, err := d.db.ListConfigs(ctx)
	if err != nil {
		return nil, translatePGError(err)
	}
	p := configuration.DBProvider{true}
	return slices.Map(configs, func(c sql.Config) configuration.Entry {
		ref := configuration.Ref{c.Module, c.Name}
		// err can be ignored because Store always returns a nil error
		u, _ := p.Store(ctx, ref, c.Value)
		return configuration.Entry{
			Ref:      ref,
			Accessor: u,
		}
	}), nil
}

func (d DatabaseResolver) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	p := configuration.DBProvider{true}
	value, err := p.Load(ctx, ref, key)
	if err != nil {
		return err
	}
	err = d.db.SetConfig(ctx, ref.Module, ref.Name, value)
	return translatePGError(err)
}

func (d DatabaseResolver) Unset(ctx context.Context, ref configuration.Ref) error {
	err := d.db.UnsetConfig(ctx, ref.Module, ref.Name)
	return translatePGError(err)
}
