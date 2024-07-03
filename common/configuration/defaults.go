package configuration

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/common/projectconfig"
)

// NewConfigurationManager creates a new configuration manager with the default configuration providers.
func NewConfigurationManager(ctx context.Context, router Router[Configuration]) (*Manager[Configuration], error) {
	return New(ctx, router, []Provider[Configuration]{
		InlineProvider[Configuration]{},
		EnvarProvider[Configuration]{},
	})
}

// NewSecretsManager creates a new secrets manager with the default secret providers.
func NewSecretsManager(ctx context.Context, router Router[Secrets], opVault string, config string) (*Manager[Secrets], error) {
	projectConfig, err := projectconfig.Load(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("could not load project config for secrets manager: %w", err)
	}
	return New(ctx, router, []Provider[Secrets]{
		InlineProvider[Secrets]{},
		EnvarProvider[Secrets]{},
		KeychainProvider{},
		OnePasswordProvider{
			Vault:       opVault,
			ProjectName: projectConfig.Name,
		},
	})
}
