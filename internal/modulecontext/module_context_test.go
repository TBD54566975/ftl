package modulecontext_test

import (
	"context" //nolint:depguard
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/log"
	. "github.com/TBD54566975/ftl/internal/modulecontext"
	. "github.com/TBD54566975/ftl/testutils/modulecontext"
)

type manualContextSupplier struct {
	initialCtx ModuleContext
	sink       func(ctx context.Context, mCtx ModuleContext)
}

func TestGettingAndSettingFromContext(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	moduleCtx := NewBuilder("test").Build()
	ctx = MakeDynamic(ctx, moduleCtx).ApplyToContext(ctx)
	assert.Equal(t, moduleCtx, FromContext(ctx).CurrentContext(), "module context should be the same when read from context")
}

func TestDynamicContextUpdate(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	mc1 := NewBuilder("test").AddConfigs(map[string][]byte{"value": {0}}).Build()
	mc2 := NewBuilder("test").AddConfigs(map[string][]byte{"value": {1}}).Build()
	mcs := &manualContextSupplier{initialCtx: mc1}
	dynamic, err := NewDynamicContext(ctx, ModuleContextSupplier(mcs), "test")
	assert.NoError(t, err)
	assert.NotEqual(t, nil, dynamic)
	assert.Equal(t, mc1, dynamic.CurrentContext())
	mcs.sink(ctx, mc2)
	assert.Equal(t, mc2, dynamic.CurrentContext())
}

func (mcs *manualContextSupplier) Subscribe(ctx context.Context, _ string, sink func(ctx context.Context, mCtx ModuleContext)) {
	sink(ctx, mcs.initialCtx)
	mcs.sink = sink
}
