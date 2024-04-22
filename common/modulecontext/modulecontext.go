package modulecontext

import (
	"context"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

// ModuleContext holds the context needed for a module, including configs, secrets and DSNs
type ModuleContext struct {
	configManager  *cf.Manager[cf.Configuration]
	secretsManager *cf.Manager[cf.Secrets]
	dbProvider     *DBProvider
}

// ApplyToContext sets up the context so that configurations, secrets and DSNs can be retreived
// Each of these components have accessors to get a manager back from the context
func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}
