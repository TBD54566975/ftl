package buildengine

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.design/x/reflect"

	"github.com/TBD54566975/ftl/common/moduleconfig"
)

// Project models FTL modules and external libraries and is used to manage dependencies within the build engine
type Project interface {
	Config() ProjectConfig
	CopyWithDependencies([]ProjectKey) Project
	String() string
}

type ProjectConfig struct {
	Key          ProjectKey
	Dir          string
	Language     string
	Watch        []string
	Dependencies []ProjectKey
}

var _ = (Project)(Module{})
var _ = (Project)(ExternalLibrary{})

// Module represents an FTL module in the build engine
type Module struct {
	moduleconfig.ModuleConfig
	Dependencies []string
}

func (m Module) Config() ProjectConfig {
	return ProjectConfig{
		Key:          ProjectKey(m.ModuleConfig.Module),
		Dir:          m.ModuleConfig.Dir,
		Language:     m.ModuleConfig.Language,
		Watch:        m.ModuleConfig.Watch,
		Dependencies: ProjectKeysFromModuleNames(m.Dependencies),
	}
}

func (m Module) CopyWithDependencies(dependencies []ProjectKey) Project {
	module := reflect.DeepCopy(m)
	module.Dependencies = StringsFromProjectKeys(dependencies)
	return Project(module)
}

func (m Module) String() string {
	return "module:" + m.ModuleConfig.Module
}

// ExternalLibrary represents a library that makes use of FTL modules, but is not itself an FTL module
type ExternalLibrary struct {
	Dir          string
	Language     string
	Dependencies []string
}

func (e ExternalLibrary) Config() ProjectConfig {
	var watch []string
	switch e.Language {
	case "go":
		watch = []string{"**/*.go", "go.mod", "go.sum"}
	case "kotlin":
		watch = []string{"pom.xml", "src/**", "target/generated-sources"}
	default:
		panic(fmt.Sprintf("unknown language %T", e.Language))
	}

	return ProjectConfig{
		Key:          ProjectKey("lib:" + e.Dir),
		Dir:          e.Dir,
		Language:     e.Language,
		Watch:        watch,
		Dependencies: ProjectKeysFromModuleNames(e.Dependencies),
	}
}

func (e ExternalLibrary) CopyWithDependencies(dependencies []ProjectKey) Project {
	lib := reflect.DeepCopy(e)
	lib.Dependencies = StringsFromProjectKeys(dependencies)
	return Project(lib)
}

func (e ExternalLibrary) String() string {
	return "library:" + e.Dir
}

// Key is a unique identifier for the project (ie: a module name or a library path)
// It is used to:
// - build the dependency graph
// - map changes in the file system to the project
type ProjectKey string

func ProjectKeyForModuleName(name string) ProjectKey {
	return ProjectKey(name)
}

func StringsFromProjectKeys(keys []ProjectKey) []string {
	strs := make([]string, len(keys))
	for i, key := range keys {
		strs[i] = string(key)
	}
	return strs
}

func ProjectKeysFromModuleNames(names []string) []ProjectKey {
	keys := make([]ProjectKey, len(names))
	for i, str := range names {
		keys[i] = ProjectKey(str)
	}
	return keys
}

// LoadModule loads a module from the given directory.
func LoadModule(dir string) (Module, error) {
	config, err := moduleconfig.LoadModuleConfig(dir)
	if err != nil {
		return Module{}, err
	}
	module := Module{ModuleConfig: config}
	return module, nil
}

func LoadExternalLibrary(dir string) (ExternalLibrary, error) {
	lib := ExternalLibrary{
		Dir: dir,
	}

	goModPath := filepath.Join(dir, "go.mod")
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(goModPath); err == nil {
		lib.Language = "go"
	} else if !os.IsNotExist(err) {
		return ExternalLibrary{}, err
	} else {
		if _, err = os.Stat(pomPath); err == nil {
			lib.Language = "kotlin"
		} else if !os.IsNotExist(err) {
			return ExternalLibrary{}, err
		}
	}
	if lib.Language == "" {
		return ExternalLibrary{}, fmt.Errorf("could not autodetect language: no go.mod or pom.xml found in %s", dir)
	}

	return lib, nil
}
