package compile

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	stdreflect "reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/TBD54566975/ftl/go-runtime/schema/finalize"
	sets "github.com/deckarep/golang-set/v2"
	gomaps "golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/reflect"
)

type ExternalModuleContext struct {
	Module         *schema.Module
	ProjectRootDir string
	GoVersion      string
	FTLVersion     string
	*schema.Schema
}

type goVerb struct {
	Name        string
	Package     string
	MustImport  string
	HasRequest  bool
	HasResponse bool
}

type mainModuleContext struct {
	GoVersion         string
	FTLVersion        string
	Name              string
	SharedModulesPath string
	Verbs             []goVerb
	Replacements      []*modfile.Replace
	SumTypes          []goSumType
}

type goWorkContext struct {
	GoVersion         string
	SharedModulesPath string
}

type goSumType struct {
	Discriminator string
	Variants      []goSumTypeVariant
}

type goSumTypeVariant struct {
	Name       string
	Type       string
	SchemaType schema.Type
}

type ModifyFilesTransaction interface {
	Begin() error
	ModifiedFiles(paths ...string) error
	End() error
}

func (b ExternalModuleContext) NonMainModules() []*schema.Module {
	// modules := make([]*schema.Module, 0, len(b.Modules))
	// for _, module := range b.Modules {
	// 	if module.Name == b.Main {
	// 		continue
	// 	}
	// 	modules = append(modules, module)
	// }
	return []*schema.Module{b.Module}
}

const buildDirName = ".ftl"

func buildDir(moduleDir string) string {
	return filepath.Join(moduleDir, buildDirName)
}

// Build the given module.
func Build(ctx context.Context, projectRootDir string, moduleConfig moduleconfig.ModuleConfig, sch *schema.Schema, filesTransaction ModifyFilesTransaction) (err error) {
	if err := filesTransaction.Begin(); err != nil {
		return err
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			err = fmt.Errorf("failed to end file transaction: %w", terr)
		}
	}()

	moduleDir := moduleConfig.Dir
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

	logger := log.FromContext(ctx)

	funcs := maps.Clone(scaffoldFuncs)

	buildDir := buildDir(moduleDir)

	err = os.Mkdir(buildDir, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	// Delete old stubs before building
	err = os.RemoveAll(filepath.Join(projectRootDir, buildDirName, "go", "modules", moduleConfig.Module))
	if err != nil {
		return err
	}

	if err := internal.ScaffoldZip(goWorkTemplateFiles(), moduleDir, goWorkContext{
		GoVersion:         goModVersion,
		SharedModulesPath: filepath.Join(projectRootDir, buildDirName, "go", "modules"),
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Debugf("Extracting schema")
	result, err := ExtractModuleSchema(moduleDir, sch)
	if err != nil {
		return err
	}

	logger.Debugf("Writing errors")
	if err = writeSchemaErrors(moduleConfig, result.Errors); err != nil {
		return fmt.Errorf("failed to write schema errors: %w", err)
	}
	if schema.ContainsTerminalError(result.Errors) {
		// Only bail if schema errors contain elements at level ERROR.
		// If errors are only at levels below ERROR (e.g. INFO, WARN), the schema can still be used.
		return nil
	}

	logger.Debugf("Writing schema")
	if err = writeSchema(moduleConfig, result.Module); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	logger.Debugf("Generating main module")
	goVerbs := make([]goVerb, 0, len(result.Module.Decls))
	for _, decl := range result.Module.Decls {
		verb, ok := decl.(*schema.Verb)
		if !ok {
			continue
		}
		nativeName, ok := result.NativeNames[verb]
		if !ok {
			return fmt.Errorf("missing native name for verb %s", verb.Name)
		}

		goverb, err := goVerbFromQualifiedName(nativeName)
		if err != nil {
			return err
		}
		if _, ok := verb.Request.(*schema.Unit); !ok {
			goverb.HasRequest = true
		}
		if _, ok := verb.Response.(*schema.Unit); !ok {
			goverb.HasResponse = true
		}
		goVerbs = append(goVerbs, goverb)
	}
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mainModuleContext{
		GoVersion:         goModVersion,
		FTLVersion:        ftlVersion,
		Name:              result.Module.Name,
		SharedModulesPath: filepath.Join(projectRootDir, buildDirName, "go", "modules"),
		Verbs:             goVerbs,
		Replacements:      replacements,
		SumTypes:          getSumTypes(result.Module, sch, result.NativeNames),
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
	if err := wg.Wait(); err != nil {
		return err
	}

	logger.Debugf("Compiling")
	return exec.Command(ctx, log.Debug, mainDir, "go", "build", "-o", "../../main", ".").RunBuffered(ctx)
}

func GenerateStubsForModule(ctx context.Context, projectRootDir string, module *schema.Module, filesTransaction ModifyFilesTransaction) (err error) {
	if err := filesTransaction.Begin(); err != nil {
		return err
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			err = fmt.Errorf("failed to end file transaction: %w", terr)
		}
	}()

	goVersion := "1.22.2" //runtime.Version()[2:]
	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	return generateExternalModules(ExternalModuleContext{
		Module:         module,
		ProjectRootDir: projectRootDir,
		Schema:         &schema.Schema{Modules: []*schema.Module{module}},
		GoVersion:      goVersion,
		FTLVersion:     ftlVersion,
	})
}

func generateExternalModules(context ExternalModuleContext) error {
	// Wipe the modules directory to ensure we don't have any stale modules.
	err := os.RemoveAll(filepath.Join(buildDir(context.ProjectRootDir), "go", "modules", context.Module.Name))
	if err != nil {
		return err
	}

	funcs := maps.Clone(scaffoldFuncs)
	return internal.ScaffoldZip(externalModuleTemplateFiles(), context.ProjectRootDir, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
}

var scaffoldFuncs = scaffolder.FuncMap{
	"comment": schema.EncodeComments,
	"type":    genType,
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

			case *schema.Topic:
				if n.IsExported() {
					imports["github.com/TBD54566975/ftl/go-runtime/ftl"] = ""
				}
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
		case *schema.TypeValue:
			return t.Value.String()
		}
		panic(fmt.Sprintf("unsupported value %T", v))
	},
	"enumInterfaceFunc": func(e schema.Enum) string {
		r := []rune(e.Name)
		for i, c := range r {
			if unicode.IsUpper(c) {
				r[i] = unicode.ToLower(c)
			} else {
				break
			}
		}
		return string(r)
	},
	"basicType": func(m *schema.Module, v schema.EnumVariant) bool {
		switch val := v.Value.(type) {
		case *schema.IntValue, *schema.StringValue:
			return false // This func should only return true for type enums
		case *schema.TypeValue:
			if _, ok := val.Value.(*schema.Ref); !ok {
				return true
			}
		}
		return false
	},
	"mainImports": func(ctx mainModuleContext) []string {
		imports := sets.NewSet[string]()
		if len(ctx.Verbs) > 0 {
			imports.Add(ctx.Name)
		}
		for _, v := range ctx.Verbs {
			imports.Add(strings.TrimPrefix(v.MustImport, "ftl/"))
		}
		for _, st := range ctx.SumTypes {
			if i := strings.LastIndex(st.Discriminator, "."); i != -1 {
				imports.Add(st.Discriminator[:i])
			}
			for _, v := range st.Variants {
				if i := strings.LastIndex(v.Type, "."); i != -1 {
					imports.Add(v.Type[:i])
				}
			}
		}
		out := imports.ToSlice()
		slices.Sort(out)
		return out
	},
	"schemaType": schemaType,
	// A standalone enum variant is one that is purely an alias to a type and does not appear
	// elsewhere in the schema.
	"isStandaloneEnumVariant": func(v schema.EnumVariant) bool {
		tv, ok := v.Value.(*schema.TypeValue)
		if !ok {
			return false
		}
		if ref, ok := tv.Value.(*schema.Ref); ok {
			return ref.Name != v.Name
		}

		return false
	},
	"sumTypes": func(m *schema.Module) []*schema.Enum {
		out := []*schema.Enum{}
		for _, d := range m.Decls {
			switch d := d.(type) {
			// Type enums (i.e. sum types) are all the non-value enums
			case *schema.Enum:
				if !d.IsValueEnum() && d.IsExported() {
					out = append(out, d)
				}
			default:
			}
		}
		return out
	},
}

func schemaType(t schema.Type) string {
	switch t := t.(type) {
	case *schema.Int, *schema.Bool, *schema.String, *schema.Float, *schema.Unit, *schema.Any, *schema.Bytes, *schema.Time:
		return fmt.Sprintf("&%s{}", strings.TrimLeft(stdreflect.TypeOf(t).String(), "*"))
	case *schema.Ref:
		return fmt.Sprintf("&schema.Ref{Module: %q, Name: %q}", t.Module, t.Name)
	case *schema.Array:
		return fmt.Sprintf("&schema.Array{Element: %s}", schemaType(t.Element))
	case *schema.Map:
		return fmt.Sprintf("&schema.Map{Key: %s, Value: %s}", schemaType(t.Key), schemaType(t.Value))
	case *schema.Optional:
		return fmt.Sprintf("&schema.Optional{Type: %s}", schemaType(t.Type))
	}
	panic(fmt.Sprintf("unsupported type %T", t))
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

func writeSchema(config moduleconfig.ModuleConfig, module *schema.Module) error {
	schemaBytes, err := proto.Marshal(module.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	return os.WriteFile(filepath.Join(config.AbsDeployDir(), "schema.pb"), schemaBytes, 0600)
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

func getSumTypes(module *schema.Module, sch *schema.Schema, nativeNames NativeNames) []goSumType {
	sumTypes := make(map[string]goSumType)
	for _, d := range module.Decls {
		if e, ok := d.(*schema.Enum); ok && !e.IsValueEnum() {
			variants := make([]goSumTypeVariant, 0, len(e.Variants))
			for _, v := range e.Variants {
				variants = append(variants, goSumTypeVariant{ //nolint:forcetypeassert
					Name:       v.Name,
					Type:       nativeNames[v],
					SchemaType: v.Value.(*schema.TypeValue).Value,
				})
			}
			stFqName := nativeNames[d]
			sumTypes[stFqName] = goSumType{
				Discriminator: nativeNames[d],
				Variants:      variants,
			}
		}
	}

	// register sum types from other modules
	for _, e := range getExternalTypeEnums(module, sch) {
		variants := make([]goSumTypeVariant, 0, len(e.resolved.Variants))
		for _, v := range e.resolved.Variants {
			variants = append(variants, goSumTypeVariant{ //nolint:forcetypeassert
				Name:       v.Name,
				Type:       e.ref.Module + "." + v.Name,
				SchemaType: v.Value.(*schema.TypeValue).Value,
			})
		}
		stFqName := e.ref.Module + "." + e.ref.Name
		sumTypes[e.ref.ToRefKey().String()] = goSumType{
			Discriminator: stFqName,
			Variants:      variants,
		}
	}
	out := gomaps.Values(sumTypes)
	slices.SortFunc(out, func(a, b goSumType) int {
		return strings.Compare(a.Discriminator, b.Discriminator)
	})
	return out
}

type externalEnum struct {
	ref      *schema.Ref
	resolved *schema.Enum
}

// getExternalTypeEnums resolve all type enum references in the full schema
func getExternalTypeEnums(module *schema.Module, sch *schema.Schema) []externalEnum {
	combinedSch := schema.Schema{
		Modules: append(sch.Modules, module),
	}
	var externalTypeEnums []externalEnum
	err := schema.Visit(&combinedSch, func(n schema.Node, next func() error) error {
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}
		if ref.Module != "" && ref.Module != module.Name {
			return next()
		}

		decl, ok := sch.Resolve(ref).Get()
		if !ok {
			return next()
		}
		if e, ok := decl.(*schema.Enum); ok && !e.IsValueEnum() {
			externalTypeEnums = append(externalTypeEnums, externalEnum{
				ref:      ref,
				resolved: e,
			})
		}
		return next()
	})
	if err != nil {
		panic(fmt.Sprintf("failed to resolve external type enums schema: %v", err))
	}
	return externalTypeEnums
}

// ExtractModuleSchema statically parses Go FTL module source into a schema.Module
//
// TODO: once migrated off of the legacy extractor, we can inline `extract.Extract(dir)` and delete this
// function
func ExtractModuleSchema(dir string, sch *schema.Schema) (finalize.Result, error) {
	result, err := extract.Extract(dir)
	if err != nil {
		return finalize.Result{}, err
	}

	// merge with legacy results for now
	if err = legacyExtractModuleSchema(dir, sch, &result); err != nil {
		return finalize.Result{}, err
	}

	schema.SortErrorsByPosition(result.Errors)
	if !schema.ContainsTerminalError(result.Errors) {
		err = schema.ValidateModule(result.Module)
		if err != nil {
			return finalize.Result{}, err
		}
	}
	updateVisibility(result.Module)
	return result, nil
}

// TODO: delete all of this once it's handled by the finalizer
func updateVisibility(module *schema.Module) {
	for _, d := range module.Decls {
		if d.IsExported() {
			updateTransitiveVisibility(d, module)
		}
	}
}

// TODO: delete
func updateTransitiveVisibility(d schema.Decl, module *schema.Module) {
	if !d.IsExported() {
		return
	}

	_ = schema.Visit(d, func(n schema.Node, next func() error) error {
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		resolved := module.Resolve(*ref)
		if resolved == nil || resolved.Symbol == nil {
			return next()
		}

		if decl, ok := resolved.Symbol.(schema.Decl); ok {
			switch t := decl.(type) {
			case *schema.Data:
				t.Export = true
			case *schema.Enum:
				t.Export = true
			case *schema.TypeAlias:
				t.Export = true
			case *schema.Topic:
				t.Export = true
			case *schema.Verb:
				t.Export = true
			case *schema.Database, *schema.Config, *schema.FSM, *schema.Secret, *schema.Subscription:
			}
		}
		return next()
	})
}

func goVerbFromQualifiedName(qualifiedName string) (goVerb, error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return goVerb{}, fmt.Errorf("invalid qualified type format %q", qualifiedName)
	}
	pkgPath := qualifiedName[:lastDotIndex]
	typeName := qualifiedName[lastDotIndex+1:]
	pkgName := path.Base(pkgPath)
	return goVerb{
		Name:       typeName,
		Package:    pkgName,
		MustImport: pkgPath,
	}, nil
}
