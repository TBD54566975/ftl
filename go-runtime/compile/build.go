package compile

import (
	"context"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
)

type buildContext struct {
	ModuleDir string
	*schema.Schema
	Main string
}

func (b buildContext) NonMainModules() []*schema.Module {
	modules := make([]*schema.Module, 0, len(b.Modules)-1)
	for _, module := range b.Modules {
		if module.Name == b.Main {
			continue
		}
		modules = append(modules, module)
	}
	return modules
}

const buildDirName = "_ftl"

// Build the given module.
func Build(ctx context.Context, moduleDir string, sch *schema.Schema) error {
	config, err := moduleconfig.LoadConfig(moduleDir)
	if err != nil {
		return fmt.Errorf("failed to load module config: %w", err)
	}
	logger := log.FromContext(ctx)

	funcs := maps.Clone(scaffoldFuncs)

	buildDir := filepath.Join(moduleDir, buildDirName)

	// Wipe the modules directory to ensure we don't have any stale modules.
	_ = os.RemoveAll(filepath.Join(buildDir, "go", "modules"))

	logger.Infof("Generating external modules")
	if err := internal.ScaffoldZip(externalModuleTemplateFiles(), moduleDir, buildContext{
		ModuleDir: moduleDir,
		Schema:    sch,
		Main:      config.Module,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Infof("Extracting schema")
	main, err := ExtractModuleSchema(moduleDir)
	if err != nil {
		return fmt.Errorf("failed to extract module schema: %w", err)
	}
	schemaBytes, err := proto.Marshal(main.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	err = os.WriteFile(filepath.Join(buildDir, "schema.pb"), schemaBytes, 0600)
	if err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	logger.Infof("Generating main module")
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, main, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Infof("Compiling")
	mainDir := filepath.Join(buildDir, "go", "main")
	if err := exec.Command(ctx, log.Debug, mainDir, "go", "mod", "tidy").Run(); err != nil {
		return fmt.Errorf("failed to tidy go.mod: %w", err)
	}
	return exec.Command(ctx, log.Info, mainDir, "go", "build", "-o", "../../main", ".").Run()
}

var scaffoldFuncs = scaffolder.FuncMap{
	"snake":          strcase.ToSnake,
	"screamingSnake": strcase.ToScreamingSnake,
	"camel":          strcase.ToCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"kebab":          strcase.ToKebab,
	"screamingKebab": strcase.ToScreamingKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename": func(v any) string {
		return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
	},
	"comment": func(s []string) string {
		if len(s) == 0 {
			return ""
		}
		return "// " + strings.Join(s, "\n// ")
	},
	// Overridden in ExternalModule().
	"type": genType,
	"is": func(kind string, t schema.Node) bool {
		return reflect.Indirect(reflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) map[string]string {
		imports := map[string]string{}
		_ = schema.Visit(m, func(n schema.Node, next func() error) error {
			switch n := n.(type) {
			case *schema.DataRef:
				if n.Module == m.Name {
					break
				}
				imports[path.Join("ftl", n.Module)] = "ftl" + n.Module

			case *schema.Time:
				imports["time"] = "stdtime"

			case *schema.Optional, *schema.Unit:
				imports["github.com/TBD54566975/ftl/go-runtime/sdk"] = ""

			default:
			}
			return next()
		})
		return imports
	},
}

func genType(module *schema.Module, t schema.Type) string {
	switch t := t.(type) {
	case *schema.DataRef:
		if module != nil && t.Module == module.Name {
			return t.Name
		}
		return "ftl" + t.String()

	case *schema.VerbRef:
		if module != nil && t.Module == module.Name {
			return t.Name
		}
		return "ftl" + t.String()

	case *schema.Float:
		return "float64"

	case *schema.Time:
		return "stdtime.Time"

	case *schema.Int, *schema.Bool, *schema.String:
		return strings.ToLower(t.String())

	case *schema.Array:
		return "[]" + genType(module, t.Element)

	case *schema.Map:
		return "map[" + genType(module, t.Key) + "]" + genType(module, t.Value)

	case *schema.Optional:
		return "sdk.Option[" + genType(module, t.Type) + "]"

	case *schema.Unit:
		return "sdk.Unit"

	case *schema.Bytes:
		return "[]byte"
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}
