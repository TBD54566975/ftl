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

	"github.com/TBD54566975/scaffolder"
	"github.com/iancoleman/strcase"
	"golang.org/x/mod/modfile"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
)

type externalModuleContext struct {
	ModuleDir string
	*schema.Schema
	GoVersion string
	Main      string
}

type mainModuleContext struct {
	GoVersion string
	*schema.Module
}

func (b externalModuleContext) NonMainModules() []*schema.Module {
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
	goModVersion, err := updateGoModule(filepath.Join(moduleDir, "go.mod"))
	if err != nil {
		return err
	}

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
	if err := internal.ScaffoldZip(externalModuleTemplateFiles(), moduleDir, externalModuleContext{
		ModuleDir: moduleDir,
		GoVersion: goModVersion,
		Schema:    sch,
		Main:      config.Module,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Infof("Tidying go.mod")
	if err := exec.Command(ctx, log.Debug, moduleDir, "go", "mod", "tidy").Run(); err != nil {
		return fmt.Errorf("failed to tidy go.mod: %w", err)
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
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mainModuleContext{
		GoVersion: goModVersion,
		Module:    main,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
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
	"typename":       schema.TypeName,
	"comment": func(s []string) string {
		if len(s) == 0 {
			return ""
		}
		return "// " + strings.Join(s, "\n// ")
	},
	"type": genType,
	"is": func(kind string, t schema.Node) bool {
		return reflect.Indirect(reflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) map[string]string {
		imports := map[string]string{}
		_ = schema.Visit(m, func(n schema.Node, next func() error) error {
			switch n := n.(type) {
			case *schema.DataRef:
				if n.Module == "" || n.Module == m.Name {
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
		desc := ""
		if module != nil && t.Module == module.Name {
			desc = t.Name
		} else if t.Module == "" {
			desc = t.Name
		} else {
			desc = "ftl" + t.Module + "." + t.Name
		}
		if len(t.TypeParameters) > 0 {
			desc += "["
			for i, tp := range t.TypeParameters {
				if i != 0 {
					desc += ", "
				}
				desc += genType(module, tp)
			}
			desc += "]"
		}
		return desc

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

	case *schema.Any:
		return "any"

	case *schema.Bytes:
		return "[]byte"

	case *schema.TypeParameter:
		return t.Name
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}

// Update go.mod file to include the FTL version.
func updateGoModule(goModPath string) (string, error) {
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}
	if !ftl.IsRelease(ftl.Version) || !shouldUpdateVersion(goModFile) {
		return goModFile.Go.Version, nil
	}
	if err := goModFile.AddRequire("github.com/TBD54566975/ftl", "v"+ftl.Version); err != nil {
		return "", fmt.Errorf("failed to add github.com/TBD54566975/ftl to %s: %w", goModPath, err)
	}

	// Atomically write the updated go.mod file.
	tmpFile, err := os.CreateTemp(filepath.Dir(goModPath), ".go.mod-")
	if err != nil {
		return "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	defer os.Remove(tmpFile.Name()) // Delete the temp file if we error.
	defer tmpFile.Close()
	goModBytes = modfile.Format(goModFile.Syntax)
	if _, err := tmpFile.Write(goModBytes); err != nil {
		return "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	if err := os.Rename(tmpFile.Name(), goModPath); err != nil {
		return "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	return goModFile.Go.Version, nil
}

func shouldUpdateVersion(goModfile *modfile.File) bool {
	for _, require := range goModfile.Require {
		if require.Mod.Path == "github.com/TBD54566975/ftl" && require.Mod.Version == ftl.Version {
			return false
		}
	}
	return true
}
