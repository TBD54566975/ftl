package buildengine

import (
	"context"
	"fmt"

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
	err := compile.GenerateStubsForModules(ctx, projectRoot, sch)
	if err != nil {
		return fmt.Errorf("failed to generate Go stubs: %w", err)
	}
	return nil
}
