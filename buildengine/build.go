package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
)

// A Module is a ModuleConfig or ExternalLibrary with its dependencies populated.
type Module struct {
	internal     interface{}
	Dependencies []string
}

type ExternalLibrary struct {
	Dir      string
	Language string
}

func (m Module) Key() string {
	switch module := m.internal.(type) {
	case moduleconfig.ModuleConfig:
		return module.Module
	case ExternalLibrary:
		return module.Dir
	default:
		panic(fmt.Sprintf("unknown internal type %T", m.internal))
	}
}

func (m Module) Language() string {
	switch module := m.internal.(type) {
	case moduleconfig.ModuleConfig:
		return module.Language
	case ExternalLibrary:
		return module.Language
	default:
		panic(fmt.Sprintf("unknown internal type %T", m.internal))
	}
}

func (m Module) Dir() string {
	switch module := m.internal.(type) {
	case moduleconfig.ModuleConfig:
		return module.Dir
	case ExternalLibrary:
		return module.Dir
	default:
		panic(fmt.Sprintf("unknown internal type %T", m.internal))
	}
}

func (m Module) Watch() []string {
	switch module := m.internal.(type) {
	case moduleconfig.ModuleConfig:
		return module.Watch
	case ExternalLibrary:
		switch module.Language {
		case "go":
			return []string{"**/*.go", "go.mod", "go.sum"}
		case "kotlin":
			return []string{"pom.xml", "src/**", "target/generated-sources"}
		default:
			panic(fmt.Sprintf("unknown language %T", m.Language()))
		}
	default:
		panic(fmt.Sprintf("unknown internal type %T", m.internal))
	}
}

func (m Module) Kind() string {
	switch m.internal.(type) {
	case moduleconfig.ModuleConfig:
		return "module"
	case ExternalLibrary:
		return "library"
	default:
		panic(fmt.Sprintf("unknown internal type %T", m.internal))
	}
}

func (m Module) ModuleConfig() (moduleconfig.ModuleConfig, bool) {
	config, ok := m.internal.(moduleconfig.ModuleConfig)
	return config, ok
}

func (m Module) ExternalLibrary() (ExternalLibrary, bool) {
	lib, ok := m.internal.(ExternalLibrary)
	return lib, ok
}

// LoadModule loads a module from the given directory.
func LoadModule(ctx context.Context, dir string) (Module, error) {
	config, err := moduleconfig.LoadModuleConfig(dir)
	if err != nil {
		return Module{}, err
	}
	module := Module{internal: config}
	return module, nil
}

func LoadExternalLibrary(ctx context.Context, dir string) (Module, error) {
	lib := ExternalLibrary{
		Dir: dir,
	}

	goModPath := filepath.Join(dir, "go.mod")
	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(goModPath); err == nil {
		lib.Language = "go"
	} else if !os.IsNotExist(err) {
		return Module{}, err
	} else {
		if _, err = os.Stat(pomPath); err == nil {
			lib.Language = "kotlin"
		} else if !os.IsNotExist(err) {
			return Module{}, err
		}
	}
	if lib.Language == "" {
		return Module{}, fmt.Errorf("could not autodetect language: no go.mod or pom.xml found in %s", dir)
	}

	module := Module{
		internal: lib,
	}
	return module, nil
}

// Build a module in the given directory given the schema and module config.
func Build(ctx context.Context, sch *schema.Schema, module Module) error {
	config, ok := module.ModuleConfig()
	if !ok {
		return fmt.Errorf("cannot build module without module config: %q", module.Key())
	}
	logger := log.FromContext(ctx).Scope(config.Module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Building module")
	switch config.Language {
	case "go":
		return buildGo(ctx, sch, module)

	case "kotlin":
		return buildKotlin(ctx, sch, module)

	default:
		return fmt.Errorf("unknown language %q", config.Language)
	}
}
