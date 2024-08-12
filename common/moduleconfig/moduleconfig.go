package moduleconfig

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/mod/modfile"

	"github.com/TBD54566975/ftl/internal/slices"
)

// ModuleGoConfig is language-specific configuration for Go modules.
type ModuleGoConfig struct{}

// ModuleKotlinConfig is language-specific configuration for Kotlin modules.
type ModuleKotlinConfig struct{}

// ModuleJavaConfig is language-specific configuration for Java modules.
type ModuleJavaConfig struct{}

// ModuleConfig is the configuration for an FTL module.
//
// Module config files are currently TOML.
type ModuleConfig struct {
	// Dir is the absolute path to the root of the module.
	Dir string `toml:"-"`

	Language string `toml:"language"`
	Realm    string `toml:"realm"`
	Module   string `toml:"module"`
	// Build is the command to build the module.
	Build string `toml:"build"`
	// Deploy is the list of files to deploy relative to the DeployDir.
	Deploy []string `toml:"deploy"`
	// DeployDir is the directory to deploy from, relative to the module directory.
	DeployDir string `toml:"deploy-dir"`
	// GeneratedSchemaDir is the directory to generate protobuf schema files into. These can be picked up by language specific build tools
	GeneratedSchemaDir string `toml:"generated-schema-dir"`
	// Schema is the name of the schema file relative to the DeployDir.
	Schema string `toml:"schema"`
	// Errors is the name of the error file relative to the DeployDir.
	Errors string `toml:"errors"`
	// Watch is the list of files to watch for changes.
	Watch []string `toml:"watch"`

	Go     ModuleGoConfig     `toml:"go,optional"`
	Kotlin ModuleKotlinConfig `toml:"kotlin,optional"`
	Java   ModuleJavaConfig   `toml:"java,optional"`
}

// AbsModuleConfig is a ModuleConfig with all paths made absolute.
//
// This is a type alias to prevent accidental use of the wrong type.
type AbsModuleConfig ModuleConfig

// LoadModuleConfig from a directory.
func LoadModuleConfig(dir string) (ModuleConfig, error) {
	path := filepath.Join(dir, "ftl.toml")
	config := ModuleConfig{}
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return ModuleConfig{}, err
	}
	if err := setConfigDefaults(dir, &config); err != nil {
		return config, fmt.Errorf("%s: %w", path, err)
	}
	config.Dir = dir
	return config, nil
}

func (c ModuleConfig) String() string {
	return fmt.Sprintf("%s (%s)", c.Module, c.Dir)
}

// Abs creates a clone of ModuleConfig with all paths made absolute.
//
// This function will panic if any paths are not beneath the module directory.
// This should never happen under normal use, as LoadModuleConfig performs this
// validation separately. This is just a sanity check.
func (c ModuleConfig) Abs() AbsModuleConfig {
	clone := c
	clone.Dir = filepath.Clean(clone.Dir)
	clone.DeployDir = filepath.Clean(filepath.Join(clone.Dir, clone.DeployDir))
	if !strings.HasPrefix(clone.DeployDir, clone.Dir) {
		panic(fmt.Sprintf("deploy-dir %q is not beneath module directory %q", clone.DeployDir, clone.Dir))
	}
	if clone.GeneratedSchemaDir != "" {
		clone.GeneratedSchemaDir = filepath.Clean(filepath.Join(clone.Dir, clone.GeneratedSchemaDir))
		if !strings.HasPrefix(clone.GeneratedSchemaDir, clone.Dir) {
			panic(fmt.Sprintf("generated-schema-dir %q is not beneath module directory %q", clone.GeneratedSchemaDir, clone.Dir))
		}
	}
	clone.Schema = filepath.Clean(filepath.Join(clone.DeployDir, clone.Schema))
	if !strings.HasPrefix(clone.Schema, clone.DeployDir) {
		panic(fmt.Sprintf("schema %q is not beneath deploy directory %q", clone.Schema, clone.DeployDir))
	}
	clone.Errors = filepath.Clean(filepath.Join(clone.DeployDir, clone.Errors))
	if !strings.HasPrefix(clone.Errors, clone.DeployDir) {
		panic(fmt.Sprintf("errors %q is not beneath deploy directory %q", clone.Errors, clone.DeployDir))
	}
	clone.Deploy = slices.Map(clone.Deploy, func(p string) string {
		out := filepath.Clean(filepath.Join(clone.DeployDir, p))
		if !strings.HasPrefix(out, clone.DeployDir) {
			panic(fmt.Sprintf("deploy path %q is not beneath deploy directory %q", out, clone.DeployDir))
		}
		return out
	})
	// Watch paths are allowed to be outside the deploy directory.
	clone.Watch = slices.Map(clone.Watch, func(p string) string {
		return filepath.Clean(filepath.Join(clone.Dir, p))
	})
	return AbsModuleConfig(clone)
}

func setConfigDefaults(moduleDir string, config *ModuleConfig) error {
	if config.Realm == "" {
		config.Realm = "home"
	}
	if config.Schema == "" {
		config.Schema = "schema.pb"
	}
	if config.Errors == "" {
		config.Errors = "errors.pb"
	}
	switch config.Language {
	case "kotlin":
		if config.Build == "" {
			config.Build = "mvn -B package"
		}
		if config.DeployDir == "" {
			config.DeployDir = "target"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "classes", "dependency", "classpath.txt"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"pom.xml", "src/**", "target/generated-sources"}
		}
	case "java":
		if config.Build == "" {
			config.Build = "mvn -B package"
		}
		if config.DeployDir == "" {
			config.DeployDir = "target"
		}
		if config.GeneratedSchemaDir == "" {
			config.GeneratedSchemaDir = "src/main/ftl-module-schema"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "quarkus-app"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"pom.xml", "src/**", "target/generated-sources"}
		}
	case "go":
		if config.DeployDir == "" {
			config.DeployDir = ".ftl"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"**/*.go", "go.mod", "go.sum"}
			watches, err := replacementWatches(moduleDir, config.DeployDir)
			if err != nil {
				return err
			}
			config.Watch = append(config.Watch, watches...)
		}

	case "rust":
		if config.Build == "" {
			config.Build = "cargo build"
		}
		if config.DeployDir == "" {
			config.DeployDir = "_ftl/target/debug"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main"}
		}
		if len(config.Watch) == 0 {
			config.Watch = []string{"**/*.rs", "Cargo.toml", "Cargo.lock"}
		}
	}

	// Do some validation.
	if !isBeneath(moduleDir, config.DeployDir) {
		return fmt.Errorf("deploy-dir must be relative to the module directory")
	}
	for _, deploy := range config.Deploy {
		if !isBeneath(moduleDir, deploy) {
			return fmt.Errorf("deploy files must be relative to the module directory")
		}
	}

	return nil
}

func isBeneath(moduleDir, path string) bool {
	resolved := filepath.Clean(filepath.Join(moduleDir, path))
	return strings.HasPrefix(resolved, strings.TrimSuffix(moduleDir, "/")+"/")
}

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
				return nil, err
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
		return nil, err
	}

	var imports []string
	for _, imp := range file.Imports {
		// Trim the quotes from the import path value
		trimmedPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, trimmedPath)
	}
	return imports, nil
}
