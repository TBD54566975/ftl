package configuration

import (
	"context"
)

// NewConfigurationManager creates a new configuration manager with the default configuration providers.
func NewConfigurationManager(ctx context.Context, resolver Resolver[Configuration]) (*Manager[Configuration], error) {
	return New(ctx, resolver, []Provider[Configuration]{
		InlineProvider[Configuration]{},
		EnvarProvider[Configuration]{},
	})
}

// NewSecretsManager creates a new secrets manager with the default secret providers.
func NewSecretsManager(ctx context.Context, resolver Resolver[Secrets], opVault string) (*Manager[Secrets], error) {
	return New(ctx, resolver, []Provider[Secrets]{
		InlineProvider[Secrets]{},
		EnvarProvider[Secrets]{},
		KeychainProvider{},
		OnePasswordProvider{Vault: opVault},
	})
}
