package dal

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/slices"
)

// DBConfigResolver loads values a project's configuration from the given database.
type DBConfigResolver struct {
	db sql.DBI
}

// DBConfigResolver should only be used for config, not secrets
var _ configuration.Resolver[configuration.Configuration] = DBConfigResolver{}

func (d DBConfigResolver) Role() configuration.Configuration { return configuration.Configuration{} }

func (d *DAL) NewConfigResolver() DBConfigResolver {
	return DBConfigResolver{db: d.db}
}

func (d DBConfigResolver) Get(ctx context.Context, ref configuration.Ref) (*url.URL, error) {
	return urlPtr(), nil
}

func (d DBConfigResolver) List(ctx context.Context) ([]configuration.Entry, error) {
	configs, err := d.db.ListConfigs(ctx)
	if err != nil {
		return nil, translatePGError(err)
	}
	return slices.Map(configs, func(c sql.Config) configuration.Entry {
		return configuration.Entry{
			Ref: configuration.Ref{
				Module: c.Module,
				Name:   c.Name,
			},
			Accessor: urlPtr(),
		}
	}), nil
}

func (d DBConfigResolver) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func (d DBConfigResolver) Unset(ctx context.Context, ref configuration.Ref) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func urlPtr() *url.URL {
	// The URLs for Database-provided configs are not actually used because all the
	// information needed to load each config is contained in the Ref, so we pass
	// around an empty "db://" to satisfy the expectations of the Resolver interface.
	return &url.URL{Scheme: "db"}
}
