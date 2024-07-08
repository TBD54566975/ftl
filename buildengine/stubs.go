package buildengine

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/go-runtime/compile"
)

// GenerateStubs generates stubs for the given modules.
//
// Currently, only Go stubs are supported. Kotlin and other language stubs can be added in the future.
func GenerateStubs(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	return generateGoStubs(ctx, projectRoot, modules, moduleConfigs)
}

// CleanStubs removes all generated stubs.
func CleanStubs(ctx context.Context, projectRoot string) error {
	return cleanGoStubs(ctx, projectRoot)
}

func generateGoStubs(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	sch := &schema.Schema{Modules: modules}
	err := compile.GenerateStubsForModules(ctx, projectRoot, moduleConfigs, sch)
	if err != nil {
		return fmt.Errorf("failed to generate go stubs: %w", err)
	}
	return nil
}

func cleanGoStubs(ctx context.Context, projectRoot string) error {
	err := compile.CleanStubs(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to clean go stubs: %w", err)
	}
	return nil
}
