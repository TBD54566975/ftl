package buildengine

import (
	"context"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

// GenerateStubs generates the stubs for the given modules.
//
// Currently, only Go is supported. Kotlin and other languages can be added here in the future.
func GenerateStubs(ctx context.Context, projectRootDir string, module *schema.Module, filesTransaction ModifyFilesTransaction) error {
	return generateGoStubs(ctx, projectRootDir, module, filesTransaction)
}

func generateGoStubs(ctx context.Context, projectRootDir string, module *schema.Module, filesTransaction ModifyFilesTransaction) error {
	err := compile.GenerateStubsForModule(ctx, projectRootDir, module, filesTransaction)
	return err
}
