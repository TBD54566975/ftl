package modulecontext

import (
	"context"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

// ModuleContext holds the context needed for a module, including configs, secrets and DSNs
type ModuleContext struct {
	module         string
	configManager  *cf.Manager[cf.Configuration]
	secretsManager *cf.Manager[cf.Secrets]
	dbProvider     *DBProvider
}

func New(module string, cm *cf.Manager[cf.Configuration], sm *cf.Manager[cf.Secrets], dbp *DBProvider) *ModuleContext {
	return &ModuleContext{
		module:         module,
		configManager:  cm,
		secretsManager: sm,
		dbProvider:     dbp,
	}
}

// ApplyToContext returns a Go context.Context that includes configurations,
// secrets and DSNs can be retreived Each of these components have accessors to
// get a manager back from the context
func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}

// FromEnvironment creates a ModuleContext from the local environment.
//
// This is useful for testing and development, where the environment is used to provide
// configurations, secrets and DSNs. The context is built from a combination of
// the ftl-project.toml file and (for now) environment variables.
func FromEnvironment(ctx context.Context, module string) (*ModuleContext, error) {
	cm, err := cf.NewDefaultConfigurationManagerFromEnvironment(ctx)
	if err != nil {
		return nil, err
	}
	sm, err := cf.NewDefaultSecretsManagerFromEnvironment(ctx)
	if err != nil {
		return nil, err
	}
	dbp, err := NewDBProviderFromEnvironment(module)
	if err != nil {
		return nil, err
	}
	return &ModuleContext{
		module:         module,
		configManager:  cm,
		secretsManager: sm,
		dbProvider:     dbp,
	}, nil
}
