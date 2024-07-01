package configuration

import (
	"context"
	"net/url"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/TBD54566975/ftl/internal/slices"
)

// DBConfigResolver loads values a project's configuration from the given database.
type DBConfigResolver struct {
	dal DBConfigResolverDAL
}

type DBConfigResolverDAL interface {
	ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error)
}

// DBConfigResolver should only be used for config, not secrets
var _ Router[Configuration] = DBConfigResolver{}

func NewDBConfigResolver(db DBConfigResolverDAL) DBConfigResolver {
	return DBConfigResolver{dal: db}
}

func (d DBConfigResolver) Role() Configuration { return Configuration{} }

func (d DBConfigResolver) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return urlPtr(), nil
}

func (d DBConfigResolver) List(ctx context.Context) ([]Entry, error) {
	configs, err := d.dal.ListModuleConfiguration(ctx)
	if err != nil {
		return nil, err
	}
	return slices.Map(configs, func(c sql.ModuleConfiguration) Entry {
		return Entry{
			Ref: Ref{
				Module: c.Module,
				Name:   c.Name,
			},
			Accessor: urlPtr(),
		}
	}), nil
}

func (d DBConfigResolver) Set(ctx context.Context, ref Ref, key *url.URL) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func (d DBConfigResolver) Unset(ctx context.Context, ref Ref) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func urlPtr() *url.URL {
	// The URLs for Database-provided configs are not actually used because all the
	// information needed to load each config is contained in the Ref, so we pass
	// around an empty "db://" to satisfy the expectations of the Resolver interface.
	return &url.URL{Scheme: "db"}
}
