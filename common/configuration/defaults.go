package configuration

import (
	"context"

	"github.com/alecthomas/kong"
)

// NewConfigurationManager constructs a new [Manager] with the default providers for configuration.
func NewConfigurationManager(ctx context.Context, configPath string) (*Manager[Configuration], error) {
	conf := DefaultConfigMixin{
		ProjectConfigResolver: ProjectConfigResolver[Configuration]{
			Config: configPath,
		},
	}
	_ = kong.ApplyDefaults(&conf)
	return conf.NewConfigurationManager(ctx)
}

// DefaultConfigMixin is a Kong mixin that provides the default configuration manager.
type DefaultConfigMixin struct {
	ProjectConfigResolver[Configuration]
	InlineProvider[Configuration]
	EnvarProvider[Configuration]
}

// NewConfigurationManager creates a new configuration manager with the default configuration providers.
func (d DefaultConfigMixin) NewConfigurationManager(ctx context.Context) (*Manager[Configuration], error) {
	return New(ctx, &d.ProjectConfigResolver, []Provider[Configuration]{
		d.InlineProvider,
		d.EnvarProvider,
	})
}

// NewSecretsManager constructs a new [Manager] with the default providers for secrets.
func NewSecretsManager(ctx context.Context, configPath string) (*Manager[Secrets], error) {
	conf := DefaultSecretsMixin{
		ProjectConfigResolver: ProjectConfigResolver[Secrets]{
			Config: configPath,
		},
	}
	_ = kong.ApplyDefaults(&conf)
	return conf.NewSecretsManager(ctx)
}

// DefaultSecretsMixin is a Kong mixin that provides the default secrets manager.
type DefaultSecretsMixin struct {
	ProjectConfigResolver[Secrets]
	InlineProvider[Secrets]
	EnvarProvider[Secrets]
	KeychainProvider
	OnePasswordProvider
}

// NewSecretsManager creates a new secrets manager with the default secret providers.
func (d DefaultSecretsMixin) NewSecretsManager(ctx context.Context) (*Manager[Secrets], error) {
	return New(ctx, &d.ProjectConfigResolver, []Provider[Secrets]{
		d.InlineProvider,
		d.EnvarProvider,
		d.KeychainProvider,
		d.OnePasswordProvider,
	})
}
