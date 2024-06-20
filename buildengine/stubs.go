package buildengine

import (
	"context"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

// GenerateStubs generates stubs for the given modules.
//
// Currently, only Go stubs are supported. Kotlin and other language stubs can be added in the future.
func GenerateStubs(ctx context.Context, projectRoot string, modules []*schema.Module) error {
	return generateGoStubs(ctx, projectRoot, modules)
}

func generateGoStubs(ctx context.Context, projectRoot string, modules []*schema.Module) error {
	sch := &schema.Schema{Modules: modules}
	return compile.GenerateStubsForModules(ctx, projectRoot, sch)
}
