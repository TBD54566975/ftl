package dal

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/alecthomas/types/optional"
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
	module, moduleOk := ref.Module.Get()
	if !moduleOk {
		module = "" // global
	}
	accessor, err := d.db.GetConfig(ctx, module, ref.Name)
	if err != nil {
		// If we could not find this config within the module, then try getting
		// again with scope set to global
		if moduleOk {
			return d.Get(ctx, configuration.Ref{
				Module: optional.None[string](),
				Name:   ref.Name,
			})
		}
		return nil, configuration.ErrNotFound
	}
	return url.Parse(accessor)
}

func (d DatabaseResolver) List(ctx context.Context) ([]configuration.Entry, error) {
	configs, err := d.db.ListConfigs(ctx)
	if err != nil {
		return nil, err
	}
	entries := []configuration.Entry{}
	for _, c := range configs {
		u, err := url.Parse(c.Accessor)
		if err != nil {
			return nil, err
		}
		entries = append(entries, configuration.Entry{
			Ref:      configuration.NewRef(c.Module, c.Name),
			Accessor: u,
		})
	}
	return entries, nil
}

func (d DatabaseResolver) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	return d.db.SetConfig(ctx, moduleAsString(ref), ref.Name, key.String())
}

func (d DatabaseResolver) Unset(ctx context.Context, ref configuration.Ref) error {
	return d.db.UnsetConfig(ctx, moduleAsString(ref), ref.Name)
}

func moduleAsString(ref configuration.Ref) string {
	module, ok := ref.Module.Get()
	if !ok {
		return ""
	}
	return module
}
