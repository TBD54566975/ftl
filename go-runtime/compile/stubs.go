package compile

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/schema"
)

type ExternalModuleContext struct {
	GoVersion    string
	FTLVersion   string
	Module       *schema.Module
	Replacements []*modfile.Replace
}

func GenerateStubs(ctx context.Context, dir string, moduleSch *schema.Module, config moduleconfig.AbsModuleConfig, nativeConfig optional.Option[moduleconfig.AbsModuleConfig]) error {
	var goModVersion string
	var replacements []*modfile.Replace
	var err error

	// If there's no module config, use the go.mod file for the first config we find.
	if config.Module == "builtin" || config.Language != "go" {
		nativeConfig, ok := nativeConfig.Get()
		if !ok {
			return fmt.Errorf("no native module config provided")
		}
		goModPath := filepath.Join(nativeConfig.Dir, "go.mod")
		_, goModVersion, err = updateGoModule(goModPath)
		if err != nil {
			return fmt.Errorf("could not read go.mod %s", goModPath)
		}
		if goModVersion == "" {
			// The best we can do here if we don't have a module to read from is to use the current Go version.
			goModVersion = runtime.Version()[2:]
		}
		replacements = []*modfile.Replace{}
	} else {
		replacements, goModVersion, err = updateGoModule(filepath.Join(config.Dir, "go.mod"))
		if err != nil {
			return err
		}
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return fmt.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	context := ExternalModuleContext{
		GoVersion:    goModVersion,
		FTLVersion:   ftlVersion,
		Module:       moduleSch,
		Replacements: replacements,
	}

	funcs := maps.Clone(scaffoldFuncs)
	err = internal.ScaffoldZip(externalModuleTemplateFiles(), dir, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
	if err != nil {
		return fmt.Errorf("failed to scaffold zip: %w", err)
	}

	if err := exec.Command(ctx, log.Debug, dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return fmt.Errorf("failed to tidy go.mod: %w", err)
	}
	return nil
}

func SyncGeneratedStubReferences(ctx context.Context, config moduleconfig.AbsModuleConfig, stubsDir string, stubbedModules []string) error {
	sharedModulePaths := []string{}
	for _, mod := range stubbedModules {
		if mod == config.Module {
			continue
		}
		sharedModulePaths = append(sharedModulePaths, filepath.Join(stubsDir, mod))
	}

	_, goModVersion, err := updateGoModule(filepath.Join(config.Dir, "go.mod"))
	if err != nil {
		return err
	}

	funcs := maps.Clone(scaffoldFuncs)
	if err := internal.ScaffoldZip(mainWorkTemplateFiles(), config.Dir, MainWorkContext{
		GoVersion:          goModVersion,
		SharedModulesPaths: sharedModulePaths,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return fmt.Errorf("failed to scaffold zip: %w", err)
	}
	return nil
}
