package modulecontext

import (
	"context"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
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
	configs    map[string][]byte
	secrets    map[string][]byte
	dsns       map[string]dsnEntry
}

func NewBuilder(moduleName string) *Builder {
	return &Builder{
		moduleName: moduleName,
		configs:    map[string][]byte{},
		secrets:    map[string][]byte{},
		dsns:       map[string]dsnEntry{},
	}
}

func NewBuilderFromProto(moduleName string, response *ftlv1.ModuleContextResponse) *Builder {
	configs := map[string][]byte{}
	for name, bytes := range response.Configs {
		configs[name] = bytes
	}
	secrets := map[string][]byte{}
	for name, bytes := range response.Secrets {
		secrets[name] = bytes
	}
	dsns := map[string]dsnEntry{}
	for _, d := range response.Databases {
		dsns[d.Name] = dsnEntry{dbType: DBType(d.Type), dsn: d.Dsn}
	}
	return &Builder{
		moduleName: moduleName,
		configs:    configs,
		secrets:    secrets,
		dsns:       dsns,
	}
}

func (b *Builder) Build(ctx context.Context) (*ModuleContext, error) {
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

	if err := buildConfigOrSecrets[cf.Configuration](ctx, *moduleCtx.configManager, b.configs, b.moduleName); err != nil {
		return nil, err
	}
	if err := buildConfigOrSecrets[cf.Secrets](ctx, *moduleCtx.secretsManager, b.secrets, b.moduleName); err != nil {
		return nil, err
	}
	for name, entry := range b.dsns {
		if err = moduleCtx.dbProvider.Add(name, entry.dbType, entry.dsn); err != nil {
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

func buildConfigOrSecrets[R cf.Role](ctx context.Context, manager cf.Manager[R], valueMap map[string][]byte, moduleName string) error {
	for name, data := range valueMap {
		if err := manager.SetData(ctx, cf.Ref{Module: optional.Some(moduleName), Name: name}, data); err != nil {
			return err
		}
	}
	return nil
}
