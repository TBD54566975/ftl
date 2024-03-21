package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGoModule(ctx context.Context, sch *schema.Schema, module Module) error {
	if err := compile.Build(ctx, module.Dir, sch); err != nil {
		return fmt.Errorf("failed to build %q: %w", module, err)
	}
	return nil
}

func buildGoLibrary(ctx context.Context, sch *schema.Schema, lib ExternalLibrary) error {
	if err := compile.GenerateStubsForExternalLibrary(ctx, lib.Dir, sch); err != nil {
		return fmt.Errorf("failed to generate stubs for %q: %w", lib, err)
	}
	return nil
}
