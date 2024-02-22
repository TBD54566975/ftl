package buildengine

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"golang.design/x/reflect"
	"golang.org/x/exp/maps"
)

// UpdateAllDependencies returns deep copies of all modules with updated dependencies.
func UpdateAllDependencies(modules []ModuleConfig) ([]ModuleConfig, error) {
	modulesByName := map[string]ModuleConfig{}
	for _, module := range modules {
		modulesByName[module.Module] = module
	}
	out := []ModuleConfig{}
	for _, module := range modules {
		updated, err := UpdateDependencies(module)
		if err != nil {
			return nil, err
		}
		out = append(out, updated)
	}
	return out, nil
}

// UpdateDependencies returns a deep copy of ModuleConfig with updated dependencies.
func UpdateDependencies(config ModuleConfig) (ModuleConfig, error) {
	dependencies, err := extractDependencies(config)
	if err != nil {
		return ModuleConfig{}, err
	}
	out := reflect.DeepCopy(config)
	out.Dependencies = dependencies
	return out, nil
}

func extractDependencies(config ModuleConfig) ([]string, error) {
	switch config.Language {
	case "go":
		return extractGoFTLImports(config.Module, config.Dir)

	case "kotlin":
		return extractKotlinFTLImports(config.Dir)

	default:
		return nil, fmt.Errorf("unsupported language: %s", config.Language)
	}
}

func extractGoFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	fset := token.NewFileSet()
	err := WalkDir(dir, func(path string, d fs.DirEntry) error {
		if !d.IsDir() {
			return nil
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
					if module == self {
						continue
					}
					dependencies[module] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Go module: %w", self, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

func extractKotlinFTLImports(dir string) ([]string, error) {
	panic("not implemented")
}
