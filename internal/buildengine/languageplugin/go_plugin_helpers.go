package languageplugin

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"

	"github.com/TBD54566975/ftl/internal/errors"
)

func replacementWatches(moduleDir, deployDir string) ([]string, error) {
	goModPath := filepath.Join(moduleDir, "go.mod")
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}

	replacements := make(map[string]string)
	for _, r := range goModFile.Replace {
		replacements[r.Old.Path] = r.New.Path
		if strings.HasPrefix(r.New.Path, ".") {
			relPath, err := filepath.Rel(filepath.Dir(goModPath), filepath.Join(filepath.Dir(goModPath), r.New.Path))
			if err != nil {
				return nil, fmt.Errorf("failed to get relative path for %s: %w", r.New.Path, err)
			}
			replacements[r.Old.Path] = relPath
		}
	}

	files, err := findReplacedImports(moduleDir, deployDir, replacements)
	if err != nil {
		return nil, err
	}

	uniquePatterns := make(map[string]struct{})
	for _, file := range files {
		pattern := filepath.Join(file, "**/*.go")
		uniquePatterns[pattern] = struct{}{}
	}

	patterns := make([]string, 0, len(uniquePatterns))
	for pattern := range uniquePatterns {
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// findReplacedImports finds Go files with imports that are specified in the replacements.
func findReplacedImports(moduleDir, deployDir string, replacements map[string]string) ([]string, error) {
	libPaths := make(map[string]bool)

	err := filepath.WalkDir(moduleDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.Contains(path, deployDir) && strings.HasSuffix(path, ".go") {
			imports, err := parseImports(path)
			if err != nil {
				return err
			}

			for _, imp := range imports {
				for oldPath, newPath := range replacements {
					if strings.HasPrefix(imp, oldPath) {
						resolvedPath := filepath.Join(newPath, strings.TrimPrefix(imp, oldPath))
						libPaths[resolvedPath] = true
						break // Only add the library path once for each import match
					}
				}
			}
		}
		return nil
	})

	return deduplicateLibPaths(libPaths), err
}

func deduplicateLibPaths(libPaths map[string]bool) []string {
	for maybeParentPath := range libPaths {
		for path := range libPaths {
			if maybeParentPath != path && strings.HasPrefix(path, maybeParentPath) {
				libPaths[path] = false
			}
		}
	}

	paths := []string{}
	for path, shouldReturn := range libPaths {
		if shouldReturn {
			paths = append(paths, path)
		}
	}

	return paths
}

func parseImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to parse imports in %s: %w", filePath, err)
	}

	var imports []string
	for _, imp := range file.Imports {
		// Trim the quotes from the import path value
		trimmedPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, trimmedPath)
	}
	return imports, nil
}
