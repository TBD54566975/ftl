package routers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/common/slices"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/dal"
)

// DatabaseConfig loads values a project's configuration from the given database.
type DatabaseConfig struct {
	dal DatabaseConfigDAL
}

type DatabaseConfigDAL interface {
	ListModuleConfiguration(ctx context.Context) ([]dal.ModuleConfiguration, error)
}

// DatabaseConfig should only be used for config, not secrets
var _ configuration.Router[configuration.Configuration] = DatabaseConfig{}

func NewDatabaseConfig(db DatabaseConfigDAL) DatabaseConfig {
	return DatabaseConfig{dal: db}
}

func (d DatabaseConfig) Role() configuration.Configuration { return configuration.Configuration{} }

func (d DatabaseConfig) Get(ctx context.Context, ref configuration.Ref) (*url.URL, error) {
	return urlPtr(), nil
}

func (d DatabaseConfig) List(ctx context.Context) ([]configuration.Entry, error) {
	configs, err := d.dal.ListModuleConfiguration(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list module configurations: %w", err)
	}
	return slices.Map(configs, func(c dal.ModuleConfiguration) configuration.Entry {
		return configuration.Entry{
			Ref: configuration.Ref{
				Module: c.Module,
				Name:   c.Name,
			},
			Accessor: urlPtr(),
		}
	}), nil
}

func (d DatabaseConfig) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func (d DatabaseConfig) Unset(ctx context.Context, ref configuration.Ref) error {
	// Writing to the DB is performed by DBConfigProvider, so this function is a NOOP
	return nil
}

func urlPtr() *url.URL {
	// The URLs for DatabaseConfig-provided configs are not actually used because all the
	// information needed to load each config is contained in the Ref, so we pass
	// around an empty "db://" to satisfy the expectations of the Resolver interface.
	return &url.URL{Scheme: "db"}
}
