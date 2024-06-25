package buildengine

import (
	"bufio"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/internal/log"
)

// UpdateDependencies finds the dependencies for a module and returns a
// Module with those dependencies populated.
func UpdateDependencies(ctx context.Context, module Module) (Module, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Extracting dependencies for %q", module.Config.Module)
	dependencies, err := extractDependencies(module)
	if err != nil {
		return Module{}, err
	}
	containsBuiltin := false
	for _, dep := range dependencies {
		if dep == "builtin" {
			containsBuiltin = true
			break
		}
	}
	if !containsBuiltin {
		dependencies = append(dependencies, "builtin")
	}

	out := module.CopyWithDependencies(dependencies)
	return out, nil
}

func extractDependencies(module Module) ([]string, error) {
	switch module.Config.Language {
	case "go":
		return extractGoFTLImports(module.Config.Module, module.Config.Dir)

	case "kotlin":
		return extractKotlinFTLImports(module.Config.Module, module.Config.Dir)

	case "rust":
		fmt.Fprintf(os.Stderr, "RUST TODO extractDependencies\n")
		return nil, nil

	default:
		return nil, fmt.Errorf("unsupported language: %s", module.Config.Language)
	}
}

func extractGoFTLImports(moduleName, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	fset := token.NewFileSet()
	err := WalkDir(dir, func(path string, d fs.DirEntry) error {
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
					if module == moduleName {
						continue
					}
					dependencies[module] = true
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Go module: %w", moduleName, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}

func extractKotlinFTLImports(self, dir string) ([]string, error) {
	dependencies := map[string]bool{}
	kotlinImportRegex := regexp.MustCompile(`^import ftl\.([A-Za-z0-9_.]+)`)

	err := filepath.WalkDir(filepath.Join(dir, "src/main/kotlin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !(strings.HasSuffix(path, ".kt") || strings.HasSuffix(path, ".kts")) {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			matches := kotlinImportRegex.FindStringSubmatch(scanner.Text())
			if len(matches) > 1 {
				module := strings.Split(matches[1], ".")[0]
				if module == self {
					continue
				}
				dependencies[module] = true
			}
		}
		return scanner.Err()
	})

	if err != nil {
		return nil, fmt.Errorf("%s: failed to extract dependencies from Kotlin module: %w", self, err)
	}
	modules := maps.Keys(dependencies)
	sort.Strings(modules)
	return modules, nil
}
