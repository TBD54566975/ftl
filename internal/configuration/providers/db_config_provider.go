package providers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/configuration"
)

// DatabaseConfig is a configuration provider that stores configuration in its key.
type DatabaseConfig struct {
	dal DatabaseConfigDAL
}

var _ configuration.SynchronousProvider[configuration.Configuration] = DatabaseConfig{}

type DatabaseConfigDAL interface {
	GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error
}

func NewDatabaseConfig(dal DatabaseConfigDAL) DatabaseConfig {
	return DatabaseConfig{
		dal: dal,
	}
}

func (DatabaseConfig) Role() configuration.Configuration { return configuration.Configuration{} }
func (DatabaseConfig) Key() string                       { return "db" }

func (d DatabaseConfig) Load(ctx context.Context, ref configuration.Ref, key *url.URL) ([]byte, error) {
	value, err := d.dal.GetModuleConfiguration(ctx, ref.Module, ref.Name)
	if err != nil {
		return nil, libdal.ErrNotFound
	}
	return value, nil
}

func (d DatabaseConfig) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	err := d.dal.SetModuleConfiguration(ctx, ref.Module, ref.Name, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set configuration: %w", err)
	}
	return &url.URL{Scheme: "db"}, nil
}

func (d DatabaseConfig) Delete(ctx context.Context, ref configuration.Ref) error {
	err := d.dal.UnsetModuleConfiguration(ctx, ref.Module, ref.Name)
	if err != nil {
		return fmt.Errorf("failed to unset configuration: %w", err)
	}
	return nil
}
