package modulecontext

import (
	"context"

	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

// ModuleContext holds the context needed for a module, including configs, secrets and DSNs
type ModuleContext struct {
	configManager  *cf.Manager[cf.Configuration]
	secretsManager *cf.Manager[cf.Secrets]
	dbProvider     *DBProvider
}

func NewFromProto(ctx context.Context, moduleName string, response *ftlv1.ModuleContextResponse) (*ModuleContext, error) {
	cm, err := newInMemoryConfigManager[cf.Configuration](ctx)
	if err != nil {
		return nil, err
	}
	sm, err := newInMemoryConfigManager[cf.Secrets](ctx)
	if err != nil {
		return nil, err
	}
	moduleCtx := &ModuleContext{
		configManager:  cm,
		secretsManager: sm,
		dbProvider:     NewDBProvider(),
	}

	if err := addConfigOrSecrets[cf.Configuration](ctx, *moduleCtx.configManager, response.Configs, moduleName); err != nil {
		return nil, err
	}
	if err := addConfigOrSecrets[cf.Secrets](ctx, *moduleCtx.secretsManager, response.Secrets, moduleName); err != nil {
		return nil, err
	}
	for _, entry := range response.Databases {
		if err = moduleCtx.dbProvider.Add(entry.Name, DBType(entry.Type), entry.Dsn); err != nil {
			return nil, err
		}
	}
	return moduleCtx, nil
}

func newInMemoryConfigManager[R cf.Role](ctx context.Context) (*cf.Manager[R], error) {
	provider := cf.NewInMemoryProvider[R]()
	resolver := cf.NewInMemoryResolver[R]()
	manager, err := cf.New(ctx, resolver, []cf.Provider[R]{provider})
	if err != nil {
		return nil, err
	}
	return manager, nil
}

func addConfigOrSecrets[R cf.Role](ctx context.Context, manager cf.Manager[R], valueMap map[string][]byte, moduleName string) error {
	for name, data := range valueMap {
		if err := manager.SetData(ctx, cf.Ref{Module: optional.Some(moduleName), Name: name}, data); err != nil {
			return err
		}
	}
	return nil
}

// ApplyToContext sets up the context so that configurations, secrets and DSNs can be retreived
// Each of these components have accessors to get a manager back from the context
func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}
