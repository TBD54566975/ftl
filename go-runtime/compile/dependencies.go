package compile

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/block/ftl/internal/moduleconfig"
	"github.com/block/ftl/internal/watch"
)

func ExtractDependencies(config moduleconfig.AbsModuleConfig) ([]string, error) {
	deps, _, err := extractDependenciesAndImports(config)
	return deps, err
}

func extractDependenciesAndImports(config moduleconfig.AbsModuleConfig) (deps []string, imports []string, err error) {
	importsMap := map[string]bool{}
	dependencies := map[string]bool{}
	fset := token.NewFileSet()
	err = watch.WalkDir(config.Dir, true, func(path string, d fs.DirEntry) error {
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
					importsMap[path] = true
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
		return nil, nil, fmt.Errorf("%s: failed to extract dependencies from Go module: %w", config.Module, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	imports = maps.Keys(importsMap)
	sort.Strings(imports)
	return modules, imports, nil
}
