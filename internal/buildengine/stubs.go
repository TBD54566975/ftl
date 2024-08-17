package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

// GenerateStubs generates stubs for the given modules.
//
// Currently, only Go stubs are supported. Kotlin and other language stubs can be added in the future.
func GenerateStubs(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	err := generateGoStubs(ctx, projectRoot, modules, moduleConfigs)
	if err != nil {
		return err
	}
	return writeGenericSchemaFiles(ctx, projectRoot, modules, moduleConfigs)
}

// CleanStubs removes all generated stubs.
func CleanStubs(ctx context.Context, projectRoot string) error {
	return cleanGoStubs(ctx, projectRoot)
}

// SyncStubReferences syncs the references in the generated stubs.
//
// For Go, this means updating all the go.work files to include all known modules in the shared stubbed modules directory.
func SyncStubReferences(ctx context.Context, projectRoot string, moduleNames []string, moduleConfigs []moduleconfig.ModuleConfig) error {
	return syncGoStubReferences(ctx, projectRoot, moduleNames, moduleConfigs)
}

func generateGoStubs(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	sch := &schema.Schema{Modules: modules}
	err := compile.GenerateStubsForModules(ctx, projectRoot, moduleConfigs, sch)
	if err != nil {
		return fmt.Errorf("failed to generate go stubs: %w", err)
	}
	return nil
}

func writeGenericSchemaFiles(ctx context.Context, projectRoot string, modules []*schema.Module, moduleConfigs []moduleconfig.ModuleConfig) error {
	sch := &schema.Schema{Modules: modules}
	for _, module := range moduleConfigs {
		if module.GeneratedSchemaDir == "" {
			continue
		}

		modPath := module.Abs().GeneratedSchemaDir
		err := os.MkdirAll(modPath, 0750)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", modPath, err)
		}

		for _, mod := range sch.Modules {
			if mod.Name == module.Module {
				continue
			}
			data, err := schema.ModuleToBytes(mod)
			if err != nil {
				return fmt.Errorf("failed to export module schema for module %s %w", mod.Name, err)
			}
			err = os.WriteFile(filepath.Join(modPath, mod.Name+".pb"), data, 0600)
			if err != nil {
				return fmt.Errorf("failed to write schema file for module %s %w", mod.Name, err)
			}
		}
	}
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

func syncGoStubReferences(ctx context.Context, projectRoot string, moduleNames []string, moduleConfigs []moduleconfig.ModuleConfig) error {
	err := compile.SyncGeneratedStubReferences(ctx, projectRoot, moduleNames, moduleConfigs)
	if err != nil {
		fmt.Printf("failed to sync go stub references: %v\n", err)
	}
	return nil
}
