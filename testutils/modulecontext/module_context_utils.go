//go:build !release

package modulecontext

import (
	"context"

	"github.com/TBD54566975/ftl/internal/modulecontext"
)

type SingleContextSupplier struct {
	moduleCtx modulecontext.ModuleContext
}

// MakeDynamic converts the specified ModuleContext to a DynamicModuleContext whose underlying
// current context never updates.
func MakeDynamic(ctx context.Context, m modulecontext.ModuleContext) *modulecontext.DynamicModuleContext {
	supplier := modulecontext.ModuleContextSupplier(SingleContextSupplier{m})
	result, err := modulecontext.NewDynamicContext(ctx, supplier, "test")
	if err != nil {
		panic(err)
	}
	return result
}

func (smc SingleContextSupplier) Subscribe(ctx context.Context, _ string, sink func(ctx context.Context, mCtx modulecontext.ModuleContext), _ func(error) bool) {
	sink(ctx, smc.moduleCtx)
}
