//go:build !release

package modulecontext

import (
	"context"
	"github.com/TBD54566975/ftl/internal/modulecontext"
)

type SingleContextSupplier struct {
	moduleCtx modulecontext.ModuleContext
}

func MakeDynamic(ctx context.Context, m modulecontext.ModuleContext) *modulecontext.DynamicModuleContext {
	supplier := modulecontext.ModuleContextSupplier(SingleContextSupplier{m})
	result, _ := modulecontext.NewDynamicContext(ctx, supplier, "test")
	return result
}

func (smc SingleContextSupplier) Subscribe(ctx context.Context, _ string, sink modulecontext.ModuleContextSink) {
	sink(ctx, smc.moduleCtx)
}
