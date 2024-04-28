package modulecontext

import (
	"context"
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/alecthomas/types/optional"
)

type dsnEntry struct {
	dbType DBType
	dsn    string
}

type valueOrData struct {
	value optional.Option[any]
	data  optional.Option[[]byte]
}

// Builder is used to set up a ModuleContext with configs, secrets and DSNs
// It is able to parse a ModuleContextResponse
type Builder struct {
	moduleName string
	configs    map[string]valueOrData
	secrets    map[string]valueOrData
	dsns       map[string]dsnEntry
}

func NewBuilder(moduleName string) *Builder {
	return &Builder{
		moduleName: moduleName,
		configs:    map[string]valueOrData{},
		secrets:    map[string]valueOrData{},
		dsns:       map[string]dsnEntry{},
	}
}

func NewBuilderFromProto(moduleName string, response *ftlv1.ModuleContextResponse) *Builder {
	configs := map[string]valueOrData{}
	for name, bytes := range response.Configs {
		configs[name] = valueOrData{data: optional.Some(bytes)}
	}
	secrets := map[string]valueOrData{}
	for name, bytes := range response.Secrets {
		secrets[name] = valueOrData{data: optional.Some(bytes)}
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

func (b *Builder) AddConfig(name string, value any) *Builder {
	b.configs[name] = valueOrData{value: optional.Some(value)}
	return b
}

func (b *Builder) AddSecret(name string, value any) *Builder {
	b.secrets[name] = valueOrData{value: optional.Some(value)}
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

func buildConfigOrSecrets[R cf.Role](ctx context.Context, manager cf.Manager[R], valueMap map[string]valueOrData, moduleName string) error {
	for name, valueOrData := range valueMap {
		if value, ok := valueOrData.value.Get(); ok {
			if err := manager.Set(ctx, cf.Ref{Module: optional.Some(moduleName), Name: name}, value); err != nil {
				return err
			}
		} else if data, ok := valueOrData.data.Get(); ok {
			if err := manager.SetData(ctx, cf.Ref{Module: optional.Some(moduleName), Name: name}, data); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("could not read value for name %q", name)
		}
	}
	return nil
}
