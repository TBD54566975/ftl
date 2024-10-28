package languageplugin

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"

	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/watch"
)

type goPlugin struct {
	*internalPlugin
}

var _ = LanguagePlugin(&goPlugin{})

func newGoPlugin(ctx context.Context) *goPlugin {
	internal := newInternalPlugin(ctx, "go", buildGo)
	return &goPlugin{
		internalPlugin: internal,
	}
}

func (p *goPlugin) ModuleConfigDefaults(ctx context.Context, dir string) (moduleconfig.CustomDefaults, error) {
	deployDir := ".ftl"
	watch := []string{"**/*.go", "go.mod", "go.sum"}
	additionalWatch, err := replacementWatches(dir, deployDir)
	watch = append(watch, additionalWatch...)
	if err != nil {
		return moduleconfig.CustomDefaults{}, err
	}
	return moduleconfig.CustomDefaults{
		Watch:     watch,
		DeployDir: deployDir,
	}, nil
}

func (p *goPlugin) GetCreateModuleFlags(ctx context.Context) ([]*kong.Flag, error) {
	return []*kong.Flag{
		{
			Value: &kong.Value{
				Name: "replace",
				Help: "Replace a module import path with a local path in the initialised FTL module.",
				Tag: &kong.Tag{
					Envs: []string{"FTL_INIT_GO_REPLACE"},
				},
			},
			Short:       'r',
			PlaceHolder: "OLD=NEW,...",
		},
	}, nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

func (p *goPlugin) CreateModule(ctx context.Context, projConfig projectconfig.Config, c moduleconfig.ModuleConfig, flags map[string]string) error {
	logger := log.FromContext(ctx)
	config := c.Abs()

	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
	}
	if !projConfig.Hermit {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}

	sctx := scaffoldingContext{
		Name:      config.Module,
		GoVersion: runtime.Version()[2:],
		Replace:   map[string]string{},
	}
	if replaceStr, ok := flags["replace"]; ok && replaceStr != "" {
		for _, replace := range strings.Split(replaceStr, ",") {
			parts := strings.Split(replace, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid replace flag (format: A=B,C=D): %q", replace)
			}
			sctx.Replace[parts[0]] = parts[1]
		}
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(config.Dir)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy: %s", config.Dir)
	if err := exec.Command(ctx, log.Debug, config.Dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return fmt.Errorf("could not tidy: %w", err)
	}
	return nil
}

func (p *goPlugin) GetDependencies(ctx context.Context, config moduleconfig.ModuleConfig) ([]string, error) {
	return p.internalPlugin.getDependencies(ctx, func() ([]string, error) {
		return compile.ExtractDependencies(config.Abs())
	})
}

func (p *goPlugin) GenerateStubs(ctx context.Context, dir string, module *schema.Module, moduleConfig moduleconfig.ModuleConfig, nativeModuleConfig optional.Option[moduleconfig.ModuleConfig]) error {
	var absNativeModuleConfig optional.Option[moduleconfig.AbsModuleConfig]
	if c, ok := nativeModuleConfig.Get(); ok {
		absNativeModuleConfig = optional.Some(c.Abs())
	}
	err := compile.GenerateStubs(ctx, dir, module, moduleConfig.Abs(), absNativeModuleConfig)
	if err != nil {
		return fmt.Errorf("could not generate stubs: %w", err)
	}
	return nil

}
func (p *goPlugin) SyncStubReferences(ctx context.Context, config moduleconfig.ModuleConfig, dir string, moduleNames []string) error {
	err := compile.SyncGeneratedStubReferences(ctx, config, dir, moduleNames)
	if err != nil {
		return fmt.Errorf("could not sync stub references: %w", err)
	}
	return nil
}

func buildGo(ctx context.Context, projectRoot, stubsRoot string, bctx BuildContext, buildEnv []string, devMode bool, transaction watch.ModifyFilesTransaction) (BuildResult, error) {
	config := bctx.Config.Abs()
	moduleSch, buildErrs, err := compile.Build(ctx, projectRoot, stubsRoot, config, bctx.Schema, transaction, buildEnv, devMode)
	if err != nil {
		return BuildResult{}, builderrors.Error{
			Msg:   "compile: " + err.Error(),
			Level: builderrors.ERROR,
			Type:  builderrors.COMPILER,
		}
	}
	return BuildResult{
		Errors: buildErrs,
		Schema: moduleSch,
		Deploy: []string{"main", "launch"},
	}, nil
}
