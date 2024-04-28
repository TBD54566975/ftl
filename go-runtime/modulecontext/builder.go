package modulecontext

import (
	"context"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/alecthomas/types/optional"
)

type dsnEntry struct {
	dbType DBType
	dsn    string
}

// Builder is used to set up a ModuleContext with configs, secrets and DSNs
// It is able to parse a ModuleContextResponse
type Builder struct {
	moduleName string
	configs    map[string]any
	secrets    map[string]any
	dsns       map[string]dsnEntry
}

func NewBuilder(moduleName string) *Builder {
	return &Builder{
		moduleName: moduleName,
		configs:    map[string]any{},
		secrets:    map[string]any{},
		dsns:       map[string]dsnEntry{},
	}
}

func (b *Builder) AddConfig(name string, value any) *Builder {
	b.configs[name] = value
	return b
}

func (b *Builder) AddSecret(name string, value any) *Builder {
	b.secrets[name] = value
	return b
}

func (b *Builder) AddDSN(name string, dbType DBType, dsn string) *Builder {
	b.dsns[name] = dsnEntry{
		dbType: dbType,
		dsn:    dsn,
	}
	return b
}

func (b *Builder) Build(ctx context.Context) (*ModuleContext, error) {
	cm, err := newInMemoryConfigManager[cf.Configuration](ctx, nil)
	if err != nil {
		return nil, err
	}
	sm, err := newInMemoryConfigManager[cf.Secrets](ctx, nil)
	if err != nil {
		return nil, err
	}
	moduleCtx := &ModuleContext{
		configManager:  cm,
		secretsManager: sm,
		dbProvider:     NewDBProvider(),
	}

	for name, value := range b.configs {
		if err := moduleCtx.configManager.Set(ctx, cf.Ref{Module: optional.Some(b.moduleName), Name: name}, value); err != nil {
			return nil, err
		}
	}
	for name, value := range b.secrets {
		if err := moduleCtx.secretsManager.Set(ctx, cf.Ref{Module: optional.Some(b.moduleName), Name: name}, value); err != nil {
			return nil, err
		}
	}
	for name, entry := range b.dsns {
		if err = moduleCtx.dbProvider.Add(name, entry.dbType, entry.dsn); err != nil {
			return nil, err
		}
	}
	return moduleCtx, nil
}
