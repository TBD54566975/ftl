package modulecontext

import (
	"context"
	"fmt"
	"os"
	"strings"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
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

func (r refValuePair) configOrSecretItem() {}

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
var _ stringResolver = envarResolver{}

type configManager[R cf.Role] struct {
	manager *cf.Manager[R]
}

func (m configManager[R]) configOrSecretItem() {}

// sumtype:decl
type configOrSecretItem interface {
	configOrSecretItem()
}

var _ configOrSecretItem = refValuePair{}
var _ configOrSecretItem = configManager[cf.Configuration]{}

type Builder struct {
	moduleName string
	configs    []configOrSecretItem
	secrets    []configOrSecretItem
	dsns       []dsnEntry
}

func NewBuilder(moduleName string) Builder {
	return Builder{
		moduleName: moduleName,
		configs:    []configOrSecretItem{},
		secrets:    []configOrSecretItem{},
		dsns:       []dsnEntry{},
	}
}

func NewBuilderFromProto(moduleName string, response *ftlv1.ModuleContextResponse) Builder {
	return Builder{
		moduleName: moduleName,
		configs: slices.Map(response.Configs, func(c *ftlv1.ModuleContextResponse_Config) configOrSecretItem {
			return refValuePair{ref: refFromProto(c.Ref), resolver: dataResolver{data: c.Data}}
		}),
		secrets: slices.Map(response.Secrets, func(s *ftlv1.ModuleContextResponse_Secret) configOrSecretItem {
			return refValuePair{ref: refFromProto(s.Ref), resolver: dataResolver{data: s.Data}}
		}),
		dsns: slices.Map(response.Databases, func(d *ftlv1.ModuleContextResponse_DSN) dsnEntry {
			return dsnEntry{name: d.Name, dbType: DBType(d.Type), resolver: valueResolver{value: d.Dsn}}
		}),
	}
}

func (b Builder) Build(ctx context.Context) (*ModuleContext, error) {
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

	if err := buildConfigOrSecrets[cf.Configuration](ctx, b, *moduleCtx.configManager, b.configs); err != nil {
		return nil, err
	}
	if err := buildConfigOrSecrets[cf.Secrets](ctx, b, *moduleCtx.secretsManager, b.secrets); err != nil {
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

func buildConfigOrSecrets[R cf.Role](ctx context.Context, b Builder, manager cf.Manager[R], items []configOrSecretItem) error {
	for _, item := range items {
		switch item := item.(type) {
		case refValuePair:
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
		case configManager[R]:
			list, err := item.manager.List(ctx)
			if err != nil {
				return err
			}
			for _, e := range list {
				if m, isModuleSpecific := e.Module.Get(); isModuleSpecific && b.moduleName != m {
					continue
				}
				data, err := item.manager.GetData(ctx, e.Ref)
				if err != nil {
					return err
				}
				if err := manager.SetData(ctx, e.Ref, data); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (b Builder) AddDSNsFromEnvarsForModule(module *schema.Module) Builder {
	// remove in favor of a non-envar approach once it is available
	for _, decl := range module.Decls {
		dbDecl, ok := decl.(*schema.Database)
		if !ok {
			continue
		}
		b.dsns = append(b.dsns, dsnEntry{
			name:   dbDecl.Name,
			dbType: DBTypePostgres,
			resolver: envarResolver{
				name: fmt.Sprintf("FTL_POSTGRES_DSN_%s_%s", strings.ToUpper(module.Name), strings.ToUpper(dbDecl.Name)),
			},
		})
	}
	return b
}

func (b Builder) AddConfigFromManager(cm *cf.Manager[cf.Configuration]) Builder {
	b.configs = append(b.configs, configManager[cf.Configuration]{manager: cm})
	return b
}

func (b Builder) AddSecretsFromManager(sm *cf.Manager[cf.Secrets]) Builder {
	b.secrets = append(b.secrets, configManager[cf.Secrets]{manager: sm})
	return b
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

type envarResolver struct {
	name string
}

func (r envarResolver) Resolve() (isData bool, value any, data []byte, err error) {
	value, err = r.ResolveString()
	if err != nil {
		return false, nil, nil, err
	}
	return false, value, nil, nil
}

func (r envarResolver) ResolveString() (string, error) {
	value, ok := os.LookupEnv(r.name)
	if !ok {
		return "", fmt.Errorf("missing environment variable %q", r.name)
	}
	return value, nil
}
