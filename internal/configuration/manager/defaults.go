package manager

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

// NewConfigurationManager creates a new configuration manager with the default configuration providers.
func NewConfigurationManager(ctx context.Context, router configuration.Router[configuration.Configuration]) (*Manager[configuration.Configuration], error) {
	return New(ctx, router, []configuration.Provider[configuration.Configuration]{
		providers.Inline[configuration.Configuration]{},
		providers.Envar[configuration.Configuration]{},
	})
}

// NewSecretsManager creates a new secrets manager with the default secret providers.
func NewSecretsManager(ctx context.Context, router configuration.Router[configuration.Secrets], opVault string, config string) (*Manager[configuration.Secrets], error) {
	projectConfig, err := projectconfig.Load(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("could not load project config for secrets manager: %w", err)
	}
	return New(ctx, router, []configuration.Provider[configuration.Secrets]{
		providers.Inline[configuration.Secrets]{},
		providers.Envar[configuration.Secrets]{},
		providers.Keychain{},
		providers.OnePassword{
			Vault:       opVault,
			ProjectName: projectConfig.Name,
		},
		providers.NewInMem[configuration.Secrets](),
	})
}
