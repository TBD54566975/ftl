package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal/log"
	"golang.org/x/mod/modfile"
)

type stubCmd struct {
	//TODO: help text not accurate
	Dir      string `arg:"" help:"Base directory containing module." type:"existingdir" required:""`
	Language string `help:"Language of the module." type:"string" required:""`
}

func (u *stubCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)

	if u.Language == "go" {
		return u.generateStubsForGoModule(ctx, logger, client)
	} else if u.Language == "kotlin" {
		//TODO: kotlin support
		panic("imlement kotlin")
	} else {
		return fmt.Errorf("language %q not supported", u.Language)
	}
}

func (u *stubCmd) generateStubsForGoModule(ctx context.Context, logger *log.Logger, client ftlv1connect.ControllerServiceClient) error {
	moduleDir := u.Dir
	buildDir := filepath.Join(moduleDir, "_ftl")

	engine, err := buildengine.New(ctx, client, []string{}, buildengine.Parallelism(1))
	if err != nil {
		return err
	}

	logger.Debugf("Remove existing build directory: %s", buildDir)
	_ = os.RemoveAll(buildDir)

	logger.Debugf("Find imported FTL modules")
	deps, err := buildengine.ExtractDependencies(moduleconfig.ModuleConfig{
		Language: "go",
		Dir:      moduleDir,
	})
	if err != nil {
		return err
	}
	logger.Debugf("Found imported FTL modules: %v", deps)

	goModPath := filepath.Join(moduleDir, "go.mod")
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}

	replacements, err := compile.PropogatedReplacements(goModFile, goModPath)
	if err != nil {
		return fmt.Errorf("failed to propogate replacements %s: %w", goModPath, err)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	modules := []*schema.Module{}
	for _, dep := range deps {
		sch, err := engine.LoadSchemaFromController(dep)
		if err != nil {
			return fmt.Errorf("failed to load schema for %s: %w", dep, err)
		}
		modules = append(modules, sch)
	}
	logger.Debugf("Found all dependent schemas")

	sch := &schema.Schema{
		Modules: modules,
	}

	err = compile.GenerateExternalModules(compile.ExternalModuleContext{
		ModuleDir:    moduleDir,
		GoVersion:    goModFile.Go.Version,
		FTLVersion:   ftlVersion,
		Schema:       sch,
		Replacements: replacements,
	})
	if err != nil {
		return fmt.Errorf("failed to generate external modules: %w", err)
	}
	logger.Infof("Generated stubs for FTL modules: %v", deps)

	return nil
}
