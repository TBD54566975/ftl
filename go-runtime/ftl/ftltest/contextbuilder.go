package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
)

type DBType int32

const (
	DBTypePostgres = DBType(modulecontext.DBTypePostgres)
)

// ContextBuilder allows for building a context with configuration, secrets and DSN values set up for tests
type ContextBuilder struct {
	builder *modulecontext.Builder
}

func NewContextBuilder(moduleName string) *ContextBuilder {
	return &ContextBuilder{
		builder: modulecontext.NewBuilder(moduleName),
	}
}

func (b *ContextBuilder) AddConfig(name string, value any) *ContextBuilder {
	b.builder.AddConfig(name, value)
	return b
}

func (b *ContextBuilder) AddSecret(name string, value any) *ContextBuilder {
	b.builder.AddSecret(name, value)
	return b
}

func (b *ContextBuilder) AddDSN(name string, dbType DBType, dsn string) *ContextBuilder {
	b.builder.AddDSN(name, modulecontext.DBType(dbType), dsn)
	return b
}

func (b *ContextBuilder) Build() (context.Context, error) {
	return b.BuildWithContext(Context())
}

func (b *ContextBuilder) BuildWithContext(ctx context.Context) (context.Context, error) {
	moduleCtx, err := b.builder.Build(ctx)
	if err != nil {
		return nil, err
	}
	return moduleCtx.ApplyToContext(ctx), nil
}
