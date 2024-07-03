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

func generateGoStubs(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	sch := &schema.Schema{Modules: modules}
	err := compile.GenerateStubsForModules(ctx, projectRoot, moduleConfigs, sch)
	if err != nil {
		return fmt.Errorf("failed to generate Go stubs: %w", err)
	}
	return nil
}
