package configuration

import (
	"context"

	"github.com/alecthomas/kong"
)

// NewConfigurationManager constructs a new [Manager] with the default providers for configuration.
func NewConfigurationManager(ctx context.Context, resolver Resolver[Configuration]) (*Manager[Configuration], error) {
	conf := DefaultConfigMixin{}
	_ = kong.ApplyDefaults(&conf)
	return conf.NewConfigurationManager(ctx, resolver)
}

// DefaultConfigMixin is a Kong mixin that provides the default configuration manager.
type DefaultConfigMixin struct {
	InlineProvider[Configuration]
	EnvarProvider[Configuration]
}

// NewConfigurationManager creates a new configuration manager with the default configuration providers.
func (d DefaultConfigMixin) NewConfigurationManager(ctx context.Context, resolver Resolver[Configuration]) (*Manager[Configuration], error) {
	return New(ctx, resolver, []Provider[Configuration]{
		d.InlineProvider,
		d.EnvarProvider,
	})
}

// NewSecretsManager constructs a new [Manager] with the default providers for secrets.
func NewSecretsManager(ctx context.Context, resolver Resolver[Secrets]) (*Manager[Secrets], error) {
	conf := DefaultSecretsMixin{}
	_ = kong.ApplyDefaults(&conf)
	return conf.NewSecretsManager(ctx, resolver)
}

// DefaultSecretsMixin is a Kong mixin that provides the default secrets manager.
type DefaultSecretsMixin struct {
	InlineProvider[Secrets]
	EnvarProvider[Secrets]
	KeychainProvider
	OnePasswordProvider
}

// NewSecretsManager creates a new secrets manager with the default secret providers.
func (d DefaultSecretsMixin) NewSecretsManager(ctx context.Context, resolver Resolver[Secrets]) (*Manager[Secrets], error) {
	return New(ctx, resolver, []Provider[Secrets]{
		d.InlineProvider,
		d.EnvarProvider,
		d.KeychainProvider,
		d.OnePasswordProvider,
	})
}
