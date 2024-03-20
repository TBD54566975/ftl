package buildengine

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TBD54566975/ftl/common/moduleconfig"
	"golang.design/x/reflect"
)

type Project interface {
	Dir() string
	Key() string
	Language() string
	Watch() []string
	Dependencies() []string
	Kind() string

	CopyWithDependencies([]string) Project
}

var _ = (Project)(Module{})
var _ = (Project)(ExternalLibrary{})

type Module struct {
	moduleconfig.ModuleConfig
	Deps []string
}

func (m Module) Dir() string {
	return m.ModuleConfig.Dir
}

func (m Module) Key() string {
	return m.ModuleConfig.Module
}

func (m Module) Language() string {
	return m.ModuleConfig.Language
}

func (m Module) Watch() []string {
	return m.ModuleConfig.Watch
}

func (m Module) Dependencies() []string {
	return m.Deps
}

func (m Module) CopyWithDependencies(dependencies []string) Project {
	module := reflect.DeepCopy(m)
	module.Deps = dependencies
	return Project(module)
}

func (m Module) Kind() string {
	return "module"
}

type ExternalLibrary struct {
	Directory string
	Lang      string
	Deps      []string
}

func (e ExternalLibrary) Dir() string {
	return e.Directory
}

func (e ExternalLibrary) Key() string {
	return e.Dir()
}

func (e ExternalLibrary) Language() string {
	return e.Lang
}

func (e ExternalLibrary) Watch() []string {
	switch e.Lang {
	case "go":
		return []string{"**/*.go", "go.mod", "go.sum"}
	case "kotlin":
		return []string{"pom.xml", "src/**", "target/generated-sources"}
	default:
		panic(fmt.Sprintf("unknown language %T", e.Lang))
	}
}

func (e ExternalLibrary) Dependencies() []string {
	return e.Deps
}

func (e ExternalLibrary) CopyWithDependencies(dependencies []string) Project {
	lib := reflect.DeepCopy(e)
	lib.Deps = dependencies
	return Project(lib)
}

func (e ExternalLibrary) Kind() string {
	return "library"
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
		Directory: dir,
	}

	goModPath := filepath.Join(dir, "go.mod")
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(goModPath); err == nil {
		lib.Lang = "go"
	} else if !os.IsNotExist(err) {
		return ExternalLibrary{}, err
	} else {
		if _, err = os.Stat(pomPath); err == nil {
			lib.Lang = "kotlin"
		} else if !os.IsNotExist(err) {
			return ExternalLibrary{}, err
		}
	}
	if lib.Lang == "" {
		return ExternalLibrary{}, fmt.Errorf("could not autodetect language: no go.mod or pom.xml found in %s", dir)
	}

	return lib, nil
}
