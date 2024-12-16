//go:build !release

package modulecontext

import (
	"context"

	"github.com/block/ftl/internal/deploymentcontext"
)

type SingleContextSupplier struct {
	moduleCtx deploymentcontext.DeploymentContext
}

// MakeDynamic converts the specified DeploymentContext to a DynamicDeploymentContext whose underlying
// current context never updates.
func MakeDynamic(ctx context.Context, m deploymentcontext.DeploymentContext) *deploymentcontext.DynamicDeploymentContext {
	supplier := deploymentcontext.DeploymentContextSupplier(SingleContextSupplier{m})
	result, err := deploymentcontext.NewDynamicContext(ctx, supplier, "test")
	if err != nil {
		panic(err)
	}
	return result
}

func (smc SingleContextSupplier) Subscribe(ctx context.Context, _ string, sink func(ctx context.Context, mCtx deploymentcontext.DeploymentContext), _ func(error) bool) {
	sink(ctx, smc.moduleCtx)
}
