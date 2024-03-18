package compile

import (
	"context"
	"fmt"
	"maps"
	"net"
	"os"
	"path"
	"path/filepath"
	stdreflect "reflect"
	"strings"

	"golang.design/x/reflect"
	"golang.org/x/mod/modfile"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

type externalModuleContext struct {
	ModuleDir string
	*schema.Schema
	GoVersion    string
	FTLVersion   string
	Main         string
	Replacements []*modfile.Replace
}

type goVerb struct {
	Name        string
	HasRequest  bool
	HasResponse bool
}

type mainModuleContext struct {
	GoVersion    string
	FTLVersion   string
	Name         string
	Verbs        []goVerb
	Replacements []*modfile.Replace
}

func (b externalModuleContext) NonMainModules() []*schema.Module {
	modules := make([]*schema.Module, 0, len(b.Modules))
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
	replacements, goModVersion, err := updateGoModule(filepath.Join(moduleDir, "go.mod"))
	if err != nil {
		return err
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	config, err := moduleconfig.LoadModuleConfig(moduleDir)
	if err != nil {
		return fmt.Errorf("failed to load module config: %w", err)
	}
	logger := log.FromContext(ctx)

	funcs := maps.Clone(scaffoldFuncs)

	buildDir := filepath.Join(moduleDir, buildDirName)

	// Wipe the modules directory to ensure we don't have any stale modules.
	_ = os.RemoveAll(filepath.Join(buildDir, "go", "modules"))

	logger.Debugf("Generating external modules")
	if err := internal.ScaffoldZip(externalModuleTemplateFiles(), moduleDir, externalModuleContext{
		ModuleDir:    moduleDir,
		GoVersion:    goModVersion,
		FTLVersion:   ftlVersion,
		Schema:       sch,
		Main:         config.Module,
		Replacements: replacements,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Debugf("Extracting schema")
	nativeNames, main, err := ExtractModuleSchema(moduleDir)
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

	logger.Debugf("Generating main module")
	goVerbs := make([]goVerb, 0, len(main.Decls))
	for _, decl := range main.Decls {
		if verb, ok := decl.(*schema.Verb); ok {
			nativeName, ok := nativeNames[verb]
			if !ok {
				return fmt.Errorf("missing native name for verb %s", verb.Name)
			}

			goverb := goVerb{Name: nativeName}
			if _, ok := verb.Request.(*schema.Unit); !ok {
				goverb.HasRequest = true
			}
			if _, ok := verb.Response.(*schema.Unit); !ok {
				goverb.HasResponse = true
			}
			goVerbs = append(goVerbs, goverb)
		}
	}
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mainModuleContext{
		GoVersion:    goModVersion,
		FTLVersion:   ftlVersion,
		Name:         main.Name,
		Verbs:        goVerbs,
		Replacements: replacements,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	wg, wgctx := errgroup.WithContext(ctx)

	logger.Debugf("Tidying go.mod files")
	wg.Go(func() error {
		if err := exec.Command(ctx, log.Debug, moduleDir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", moduleDir, err)
		}
		return nil
	})
	mainDir := filepath.Join(buildDir, "go", "main")
	wg.Go(func() error {
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", mainDir, err)
		}
		return nil
	})
	wg.Go(func() error {
		modulesDir := filepath.Join(buildDir, "go", "modules")
		if err := exec.Command(wgctx, log.Debug, modulesDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", modulesDir, err)
		}
		return nil
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	logger.Debugf("Compiling")
	return exec.Command(ctx, log.Debug, mainDir, "go", "build", "-o", "../../main", ".").RunBuffered(ctx)
}

func online() bool {
	_, err := net.LookupHost("proxy.golang.org")
	return err == nil
}

var scaffoldFuncs = scaffolder.FuncMap{
	"comment": func(s []string) string {
		if len(s) == 0 {
			return ""
		}
		return "// " + strings.Join(s, "\n// ")
	},
	"type": genType,
	"is": func(kind string, t schema.Node) bool {
		return stdreflect.Indirect(stdreflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) map[string]string {
		imports := map[string]string{}
		_ = schema.Visit(m, func(n schema.Node, next func() error) error {
			switch n := n.(type) {
			case *schema.Ref:
				if n.Module == "" || n.Module == m.Name {
					break
				}
				imports[path.Join("ftl", n.Module)] = "ftl" + n.Module

				for _, tp := range n.TypeParameters {
					tpRef, err := schema.ParseRef(tp.String())
					if err != nil {
						return err
					}
					if tpRef.Module != "" && tpRef.Module != m.Name {
						imports[path.Join("ftl", tpRef.Module)] = "ftl" + tpRef.Module
					}
				}

			case *schema.Time:
				imports["time"] = "stdtime"

			case *schema.Optional, *schema.Unit:
				imports["github.com/TBD54566975/ftl/go-runtime/ftl"] = ""

			default:
			}
			return next()
		})
		return imports
	},
	"value": func(v schema.Value) string {
		switch t := v.(type) {
		case *schema.StringValue:
			return fmt.Sprintf("%q", t.Value)
		case *schema.IntValue:
			return fmt.Sprintf("%d", t.Value)
		}
		panic(fmt.Sprintf("unsupported value %T", v))
	},
}

func genType(module *schema.Module, t schema.Type) string {
	switch t := t.(type) {
	case *schema.Ref:
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
		return "ftl.Option[" + genType(module, t.Type) + "]"

	case *schema.Unit:
		return "ftl.Unit"

	case *schema.Any:
		return "any"

	case *schema.Bytes:
		return "[]byte"
	}
	panic(fmt.Sprintf("unsupported type %T", t))
}

// Update go.mod file to include the FTL version and return the Go version and any replace directives.
func updateGoModule(goModPath string) (replacements []*modfile.Replace, goVersion string, err error) {
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}

	// Propagate any replace directives.
	replacements = reflect.DeepCopy(goModFile.Replace)
	for i, r := range replacements {
		if strings.HasPrefix(r.New.Path, ".") {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(goModPath), r.New.Path))
			if err != nil {
				return nil, "", err
			}
			replacements[i].New.Path = abs
		}
	}

	// Early return if we're not updating anything.
	if !ftl.IsRelease(ftl.Version) || !shouldUpdateVersion(goModFile) {
		return replacements, goModFile.Go.Version, nil
	}

	if err := goModFile.AddRequire("github.com/TBD54566975/ftl", "v"+ftl.Version); err != nil {
		return nil, "", fmt.Errorf("failed to add github.com/TBD54566975/ftl to %s: %w", goModPath, err)
	}

	// Atomically write the updated go.mod file.
	tmpFile, err := os.CreateTemp(filepath.Dir(goModPath), ".go.mod-")
	if err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	defer os.Remove(tmpFile.Name()) // Delete the temp file if we error.
	defer tmpFile.Close()
	goModBytes = modfile.Format(goModFile.Syntax)
	if _, err := tmpFile.Write(goModBytes); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	if err := os.Rename(tmpFile.Name(), goModPath); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	return replacements, goModFile.Go.Version, nil
}

func shouldUpdateVersion(goModfile *modfile.File) bool {
	for _, require := range goModfile.Require {
		if require.Mod.Path == "github.com/TBD54566975/ftl" && require.Mod.Version == ftl.Version {
			return false
		}
	}
	return true
}
