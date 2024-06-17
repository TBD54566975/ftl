//go:build !release

package modulecontext

import (
	"context"
)

type SingleContextSupplier struct {
	moduleCtx ModuleContext
}

func (m ModuleContext) MakeDynamic(ctx context.Context) *DynamicModuleContext {
	supplier := ModuleContextSupplier(SingleContextSupplier{m})
	result, _ := NewDynamicContext(ctx, supplier, "test")
	return result
}

func (smc SingleContextSupplier) Subscribe(ctx context.Context, _ string, sink ModuleContextSink) {
	sink(ctx, smc.moduleCtx)
}
