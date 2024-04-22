package modulecontext

import (
	"context"
	"fmt"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
)

type ref struct {
	Module optional.Option[string]
	Name   string
}

type refValuePair struct {
	ref      ref
	resolver resolver
}

type dsnEntry struct {
	name     string
	dbType   DBType
	resolver stringResolver
}

type resolver interface {
	Resolve() (isData bool, value any, data []byte, err error)
}

var _ resolver = valueResolver{}
var _ resolver = dataResolver{}

type stringResolver interface {
	ResolveString() (string, error)
}

var _ stringResolver = valueResolver{}

type Builder struct {
	moduleName string
	configs    []refValuePair
	secrets    []refValuePair
	dsns       []dsnEntry
}

func NewBuilder(moduleName string) *Builder {
	return &Builder{
		moduleName: moduleName,
		configs:    []refValuePair{},
		secrets:    []refValuePair{},
		dsns:       []dsnEntry{},
	}
}

func NewBuilderFromProto(moduleName string, response *ftlv1.ModuleContextResponse) *Builder {
	return &Builder{
		moduleName: moduleName,
		configs: slices.Map(response.Configs, func(c *ftlv1.ModuleContextResponse_Config) refValuePair {
			return refValuePair{ref: refFromProto(c.Ref), resolver: dataResolver{data: c.Data}}
		}),
		secrets: slices.Map(response.Secrets, func(s *ftlv1.ModuleContextResponse_Secret) refValuePair {
			return refValuePair{ref: refFromProto(s.Ref), resolver: dataResolver{data: s.Data}}
		}),
		dsns: slices.Map(response.Databases, func(d *ftlv1.ModuleContextResponse_DSN) dsnEntry {
			return dsnEntry{name: d.Name, dbType: DBType(d.Type), resolver: valueResolver{value: d.Dsn}}
		}),
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

	if err := buildConfigOrSecrets[cf.Configuration](ctx, *moduleCtx.configManager, b.configs); err != nil {
		return nil, err
	}
	if err := buildConfigOrSecrets[cf.Secrets](ctx, *moduleCtx.secretsManager, b.secrets); err != nil {
		return nil, err
	}

	for _, entry := range b.dsns {
		dsn, err := entry.resolver.ResolveString()
		if err != nil {
			return nil, err
		}
		if err = moduleCtx.dbProvider.AddDSN(entry.name, entry.dbType, dsn); err != nil {
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

func buildConfigOrSecrets[R cf.Role](ctx context.Context, manager cf.Manager[R], items []refValuePair) error {
	for _, item := range items {
		isData, value, data, err := item.resolver.Resolve()
		if err != nil {
			return err
		}
		if isData {
			if err := manager.SetData(ctx, cf.Ref(item.ref), data); err != nil {
				return err
			}
		} else {
			if err := manager.Set(ctx, cf.Ref(item.ref), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func refFromProto(r *ftlv1.ModuleContextResponse_Ref) ref {
	return ref{
		Module: optional.Ptr(r.Module),
		Name:   r.Name,
	}
}

type valueResolver struct {
	value any
}

func (r valueResolver) Resolve() (isData bool, value any, data []byte, err error) {
	return false, r.value, nil, nil
}

func (r valueResolver) ResolveString() (string, error) {
	str, ok := r.value.(string)
	if !ok {
		return "", fmt.Errorf("value is not a string: %v", r.value)
	}
	return str, nil
}

type dataResolver struct {
	data []byte
}

func (r dataResolver) Resolve() (isData bool, value any, data []byte, err error) {
	return true, nil, r.data, nil
}
