package languageplugin

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/kong"
	"golang.org/x/exp/maps"

	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal"
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
		Deploy:    []string{"main", "launch"},
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
		dependencies := map[string]bool{}
		fset := token.NewFileSet()
		err := watch.WalkDir(config.Abs().Dir, func(path string, d fs.DirEntry) error {
			if !d.IsDir() {
				return nil
			}
			if strings.HasPrefix(d.Name(), "_") || d.Name() == "testdata" {
				return watch.ErrSkip
			}
			pkgs, err := parser.ParseDir(fset, path, nil, parser.ImportsOnly)
			if pkgs == nil {
				return fmt.Errorf("could parse directory in search of dependencies: %w", err)
			}
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					for _, imp := range file.Imports {
						path, err := strconv.Unquote(imp.Path.Value)
						if err != nil {
							continue
						}
						if !strings.HasPrefix(path, "ftl/") {
							continue
						}
						module := strings.Split(strings.TrimPrefix(path, "ftl/"), "/")[0]
						if module == config.Module {
							continue
						}
						dependencies[module] = true
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("%s: failed to extract dependencies from Go module: %w", config.Module, err)
		}
		modules := maps.Keys(dependencies)
		sort.Strings(modules)
		return modules, nil
	})
}

func buildGo(ctx context.Context, projectRoot string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction watch.ModifyFilesTransaction) error {
	if err := compile.Build(ctx, projectRoot, config.Dir, config, sch, transaction, buildEnv, devMode); err != nil {
		return CompilerBuildError{err: fmt.Errorf("failed to build module %q: %w", config.Module, err)}
	}
	return nil
}
