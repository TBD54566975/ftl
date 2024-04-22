package modulecontext

import (
	"context"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

type ModuleContext struct {
	configManager  *cf.Manager[cf.Configuration]
	secretsManager *cf.Manager[cf.Secrets]
	dbProvider     *DBProvider
}

func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}
