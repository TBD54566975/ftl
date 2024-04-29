package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
)

type DBType int32

const (
	DBTypePostgres = DBType(modulecontext.DBTypePostgres)
)

// ContextBuilder allows for building a context with configuration, secrets and DSN values set up for tests
type ContextBuilder struct {
	builder *modulecontext.Builder

	mocks map[ftl.Ref]mockFunc
}

func NewContextBuilder(moduleName string) *ContextBuilder {
	return &ContextBuilder{
		builder: modulecontext.NewBuilder(moduleName),
		mocks:   map[ftl.Ref]mockFunc{},
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

func (b *ContextBuilder) MockVerb(moduleName string, name string, mock func(ctx context.Context, req any) (any, error)) *ContextBuilder {
	b.mocks[ftl.Ref{Module: moduleName, Name: name}] = mock
	return b
}

func (b *ContextBuilder) MockSinkVerb(moduleName string, name string, mock func(ctx context.Context, req any) error) *ContextBuilder {
	b.mocks[ftl.Ref{Module: moduleName, Name: name}] = func(ctx context.Context, req any) (any, error) {
		return ftl.Unit{}, mock(ctx, req)
	}
	return b
}

func (b *ContextBuilder) MockSourceVerb(moduleName string, name string, mock func(ctx context.Context) (any, error)) *ContextBuilder {
	b.mocks[ftl.Ref{Module: moduleName, Name: name}] = func(ctx context.Context, req any) (any, error) {
		return mock(ctx)
	}
	return b
}

func (b *ContextBuilder) MockEmptyVerb(moduleName string, name string, mock func(ctx context.Context) error) *ContextBuilder {
	b.mocks[ftl.Ref{Module: moduleName, Name: name}] = func(ctx context.Context, req any) (any, error) {
		return ftl.Unit{}, mock(ctx)
	}
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

	ctx = moduleCtx.ApplyToContext(ctx)

	mockProvider := NewMockProvider(ctx, b.mocks)
	ctx = ftl.ApplyCallOverriderToContext(ctx, mockProvider)

	return ctx, nil
}
