package buildengine

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
	"github.com/alecthomas/types/pubsub"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
)

type goPlugin struct {
	*internalPlugin
}

var _ = Plugin(&goPlugin{})

func newGoPlugin(ctx context.Context, config moduleconfig.AbsModuleConfig, projectPath string) *goPlugin {
	internal := newInternalPlugin(ctx, config, func(ctx context.Context, config moduleconfig.AbsModuleConfig, sch *schema.Schema, buildEnv []string, devMode bool, transaction ModifyFilesTransaction) error {
		if err := compile.Build(ctx, projectPath, config.Dir, sch, transaction, buildEnv, devMode); err != nil {
			return CompilerBuildError{err: fmt.Errorf("failed to build module %q: %w", config.Module, err)}
		}
		return nil
	})
	return &goPlugin{
		internalPlugin: internal,
	}
}

func (p *goPlugin) Updates() *pubsub.Topic[PluginEvent] {
	return p.internalPlugin.updates
}

func (p *goPlugin) Kill(ctx context.Context) error {
	// TODO: create own context for background execution and cancel that...
	return nil
}

type scaffoldingContext struct {
	Name      string
	GoVersion string
	Replace   map[string]string
}

func (p *goPlugin) CreateModule(ctx context.Context, config moduleconfig.AbsModuleConfig) error {
	logger := log.FromContext(ctx)
	opts := []scaffolder.Option{
		scaffolder.Exclude("^go.mod$"),
		scaffolder.Functions(scaffoldFuncs),
	}
	// TODO: bring back this logic
	// if !includeBinDir {
	logger.Debugf("Excluding bin directory")
	opts = append(opts, scaffolder.Exclude("^bin"))
	// }
	sctx := scaffoldingContext{
		Name:      config.Module,
		GoVersion: runtime.Version()[2:],
	}

	// scaffold at one directory above the module directory
	parentPath := filepath.Dir(config.Dir)
	if err := internal.ScaffoldZip(goruntime.Files(), parentPath, sctx, opts...); err != nil {
		return fmt.Errorf("failed to scaffold: %w", err)
	}
	logger.Debugf("Running go mod tidy: %s", config.Dir)
	if err := exec.Command(ctx, log.Debug, config.Dir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
		return err
	}
	return nil
}

func (p *goPlugin) GetDependencies(ctx context.Context) ([]string, error) {
	return p.internalPlugin.getDependencies(ctx, func() ([]string, error) {
		dependencies := map[string]bool{}
		fset := token.NewFileSet()
		err := WalkDir(p.config.Dir, func(path string, d fs.DirEntry) error {
			if !d.IsDir() {
				return nil
			}
			if strings.HasPrefix(d.Name(), "_") || d.Name() == "testdata" {
				return ErrSkip
			}
			pkgs, err := parser.ParseDir(fset, path, nil, parser.ImportsOnly)
			if pkgs == nil {
				return err
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
						if module == p.config.Module {
							continue
						}
						dependencies[module] = true
					}
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("%s: failed to extract dependencies from Go module: %w", p.config.Module, err)
		}
		modules := maps.Keys(dependencies)
		sort.Strings(modules)
		return modules, nil
	})
}

func (p *goPlugin) Build(ctx context.Context, config moduleconfig.AbsModuleConfig, sch *schema.Schema, projectPath string, buildEnv []string, devMode bool) (BuildResult, error) {
	return p.internalPlugin.build(ctx, config, sch, buildEnv, devMode)
}
