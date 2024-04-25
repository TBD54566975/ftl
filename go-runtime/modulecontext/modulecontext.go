package modulecontext

import (
	"context"
	"net/url"

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
	cm, err := newInMemoryConfigManager[cf.Configuration](ctx, response.Configs)
	if err != nil {
		return nil, err
	}
	sm, err := newInMemoryConfigManager[cf.Secrets](ctx, response.Secrets)
	if err != nil {
		return nil, err
	}
	moduleCtx := &ModuleContext{
		configManager:  cm,
		secretsManager: sm,
		dbProvider:     NewDBProvider(),
	}

	for _, entry := range response.Databases {
		if err = moduleCtx.dbProvider.Add(entry.Name, DBType(entry.Type), entry.Dsn); err != nil {
			return nil, err
		}
	}
	return moduleCtx, nil
}

func newInMemoryConfigManager[R cf.Role](ctx context.Context, config map[string][]byte) (*cf.Manager[R], error) {
	provider := cf.NewInMemoryProvider[R]()
	refs := map[cf.Ref]*url.URL{}
	for name, data := range config {
		ref := cf.Ref{Name: name}
		u, err := provider.Store(ctx, ref, data)
		if err != nil {
			return nil, err
		}
		refs[ref] = u
	}
	resolver := cf.NewInMemoryResolver[R]()
	for ref, u := range refs {
		err := resolver.Set(ctx, ref, u)
		if err != nil {
			return nil, err
		}
	}
	manager, err := cf.New(ctx, resolver, []cf.Provider[R]{provider})
	if err != nil {
		return nil, err
	}
	return manager, nil
}

// ApplyToContext sets up the context so that configurations, secrets and DSNs can be retreived
// Each of these components have accessors to get a manager back from the context
func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}
