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
	"runtime"
	"strconv"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/reflect"
)

type ExternalModuleContext struct {
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

type ModifyFilesTransaction interface {
	Begin() error
	ModifiedFiles(paths ...string) error
	End() error
}

func (b ExternalModuleContext) NonMainModules() []*schema.Module {
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

func buildDir(moduleDir string) string {
	return filepath.Join(moduleDir, buildDirName)
}

// Build the given module.
func Build(ctx context.Context, moduleDir string, sch *schema.Schema, filesTransaction ModifyFilesTransaction) error {
	if err := filesTransaction.Begin(); err != nil {
		return err
	}
	defer func() {
		if err := filesTransaction.End(); err != nil {
			log.FromContext(ctx).Errorf(err, "failed to end file transaction")
		}
	}()

	replacements, goModVersion, err := updateGoModule(filepath.Join(moduleDir, "go.mod"))
	if err != nil {
		return err
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return fmt.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
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

	logger.Debugf("Generating external modules")
	if err := generateExternalModules(ExternalModuleContext{
		ModuleDir:    moduleDir,
		GoVersion:    goModVersion,
		FTLVersion:   ftlVersion,
		Schema:       sch,
		Main:         config.Module,
		Replacements: replacements,
	}); err != nil {
		return fmt.Errorf("failed to generate external modules: %w", err)
	}

	buildDir := buildDir(moduleDir)
	logger.Debugf("Extracting schema")
	nativeNames, main, schemaErrs, err := ExtractModuleSchema(moduleDir)
	if err != nil {
		return fmt.Errorf("failed to extract module schema: %w", err)
	}
	if len(schemaErrs) > 0 {
		if err := writeSchemaErrors(config, schemaErrs); err != nil {
			return fmt.Errorf("failed to write errors: %w", err)
		}
		return nil
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

	logger.Debugf("Tidying go.mod files")
	wg, wgctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := exec.Command(ctx, log.Debug, moduleDir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", moduleDir, err)
		}
		return filesTransaction.ModifiedFiles(filepath.Join(moduleDir, "go.mod"), filepath.Join(moduleDir, "go.sum"))
	})
	mainDir := filepath.Join(buildDir, "go", "main")
	wg.Go(func() error {
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", mainDir, err)
		}
		return filesTransaction.ModifiedFiles(filepath.Join(mainDir, "go.mod"), filepath.Join(moduleDir, "go.sum"))
	})
	wg.Go(func() error {
		modulesDir := filepath.Join(buildDir, "go", "modules")
		if err := exec.Command(wgctx, log.Debug, modulesDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", modulesDir, err)
		}
		return filesTransaction.ModifiedFiles(filepath.Join(modulesDir, "go.mod"), filepath.Join(moduleDir, "go.sum"))
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	logger.Debugf("Compiling")
	err = exec.Command(ctx, log.Debug, mainDir, "go", "build", "-o", "../../main", ".").RunBuffered(ctx)
	if err != nil {
		return err
	}

	return nil
}

func GenerateStubsForExternalLibrary(ctx context.Context, dir string, schema *schema.Schema) error {
	goModFile, replacements, err := goModFileWithReplacements(filepath.Join(dir, "go.mod"))
	if err != nil {
		return fmt.Errorf("failed to propagate replacements for library %q: %w", dir, err)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	return generateExternalModules(ExternalModuleContext{
		ModuleDir:    dir,
		GoVersion:    goModFile.Go.Version,
		FTLVersion:   ftlVersion,
		Schema:       schema,
		Replacements: replacements,
	})

}

func generateExternalModules(context ExternalModuleContext) error {
	// Wipe the modules directory to ensure we don't have any stale modules.
	err := os.RemoveAll(filepath.Join(buildDir(context.ModuleDir), "go", "modules"))
	if err != nil {
		return err
	}

	funcs := maps.Clone(scaffoldFuncs)
	return internal.ScaffoldZip(externalModuleTemplateFiles(), context.ModuleDir, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
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
		_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error {
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
			return strconv.Itoa(t.Value)
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
	goModFile, replacements, err := goModFileWithReplacements(goModPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update %s: %w", goModPath, err)
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
	goModBytes := modfile.Format(goModFile.Syntax)
	if _, err := tmpFile.Write(goModBytes); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	if err := os.Rename(tmpFile.Name(), goModPath); err != nil {
		return nil, "", fmt.Errorf("update %s: %w", goModPath, err)
	}
	return replacements, goModFile.Go.Version, nil
}

func goModFileWithReplacements(goModPath string) (*modfile.File, []*modfile.Replace, error) {
	goModBytes, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s: %w", goModPath, err)
	}
	goModFile, err := modfile.Parse(goModPath, goModBytes, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s: %w", goModPath, err)
	}

	replacements := reflect.DeepCopy(goModFile.Replace)
	for i, r := range replacements {
		if strings.HasPrefix(r.New.Path, ".") {
			abs, err := filepath.Abs(filepath.Join(filepath.Dir(goModPath), r.New.Path))
			if err != nil {
				return nil, nil, err
			}
			replacements[i].New.Path = abs
		}
	}
	return goModFile, replacements, nil
}

func shouldUpdateVersion(goModfile *modfile.File) bool {
	for _, require := range goModfile.Require {
		if require.Mod.Path == "github.com/TBD54566975/ftl" && require.Mod.Version == ftl.Version {
			return false
		}
	}
	return true
}

func writeSchemaErrors(config moduleconfig.ModuleConfig, errors []*schema.Error) error {
	el := schema.ErrorList{
		Errors: errors,
	}
	elBytes, err := proto.Marshal(el.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}
	return os.WriteFile(filepath.Join(config.AbsDeployDir(), config.Errors), elBytes, 0600)
}
