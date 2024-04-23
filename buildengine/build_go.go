package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

func buildGoModule(ctx context.Context, sch *schema.Schema, module Module, transaction ModifyFilesTransaction) error {
	if err := compile.Build(ctx, module.Dir, sch, transaction); err != nil {
		return fmt.Errorf("failed to build module %q: %w", module.Config().Key, err)
	}
	return nil
}

func buildGoLibrary(ctx context.Context, sch *schema.Schema, lib ExternalLibrary) error {
	if err := compile.GenerateStubsForExternalLibrary(ctx, lib.Dir, sch); err != nil {
		return fmt.Errorf("failed to generate stubs for library %q: %w", lib.Config().Key, err)
	}
	return nil
}
