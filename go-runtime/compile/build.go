package compile

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	stdreflect "reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/reflect"
)

type MainWorkContext struct {
	GoVersion          string
	SharedModulesPaths []string
}

type ExternalModuleContext struct {
	*schema.Schema
	GoVersion    string
	FTLVersion   string
	Module       *schema.Module
	Replacements []*modfile.Replace
}

type goVerb struct {
	Name        string
	Package     string
	MustImport  string
	HasRequest  bool
	HasResponse bool
}

type mainModuleContext struct {
	GoVersion          string
	FTLVersion         string
	Name               string
	SharedModulesPaths []string
	Verbs              []goVerb
	Replacements       []*modfile.Replace
	SumTypes           []goSumType
	LocalSumTypes      []goSumType
	ExternalTypes      []goExternalType
	LocalExternalTypes []goExternalType
}

type goSumType struct {
	Discriminator string
	Variants      []goSumTypeVariant
	fqName        string
}

type goSumTypeVariant struct {
	Name       string
	Type       string
	SchemaType schema.Type
}

type goExternalType struct {
	Import string
	Types  []string
}

type ModifyFilesTransaction interface {
	Begin() error
	ModifiedFiles(paths ...string) error
	End() error
}

const buildDirName = ".ftl"

func buildDir(moduleDir string) string {
	return filepath.Join(moduleDir, buildDirName)
}

// Build the given module.
func Build(ctx context.Context, projectRootDir, moduleDir string, sch *schema.Schema, filesTransaction ModifyFilesTransaction) (err error) {
	if err := filesTransaction.Begin(); err != nil {
		return err
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			err = fmt.Errorf("failed to end file transaction: %w", terr)
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

	buildDir := buildDir(moduleDir)
	err = os.MkdirAll(buildDir, 0750)
	if err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	var sharedModulesPaths []string
	for _, mod := range sch.Modules {
		if mod.Name == config.Module {
			continue
		}
		sharedModulesPaths = append(sharedModulesPaths, filepath.Join(projectRootDir, buildDirName, "go", "modules", mod.Name))
	}

	if err := internal.ScaffoldZip(mainWorkTemplateFiles(), moduleDir, MainWorkContext{
		GoVersion:          goModVersion,
		SharedModulesPaths: sharedModulesPaths,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return fmt.Errorf("failed to scaffold zip: %w", err)
	}

	logger.Debugf("Extracting schema")
	result, err := ExtractModuleSchema(config.Dir, sch)
	if err != nil {
		return err
	}

	if err = writeSchemaErrors(config, result.Errors); err != nil {
		return fmt.Errorf("failed to write schema errors: %w", err)
	}
	if schema.ContainsTerminalError(result.Errors) {
		// Only bail if schema errors contain elements at level ERROR.
		// If errors are only at levels below ERROR (e.g. INFO, WARN), the schema can still be used.
		return nil
	}
	if err = writeSchema(config, result.Module); err != nil {
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
	localExternalTypes, err := getLocalExternalTypes(result.Module)
	if err != nil {
		return err
	}
	allSumTypes, allExternalTypes, err := getRegisteredTypes(result.Module, sch, result.NativeNames)
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mainModuleContext{
		GoVersion:          goModVersion,
		FTLVersion:         ftlVersion,
		Name:               result.Module.Name,
		SharedModulesPaths: sharedModulesPaths,
		Verbs:              goVerbs,
		Replacements:       replacements,
		SumTypes:           allSumTypes,
		LocalSumTypes:      getLocalSumTypes(result.Module, result.NativeNames),
		ExternalTypes:      allExternalTypes,
		LocalExternalTypes: localExternalTypes,
	}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
		return err
	}

	logger.Debugf("Tidying go.mod files")
	wg, wgctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := exec.Command(wgctx, log.Debug, moduleDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
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

// CleanStubs removes all generated stubs.
func CleanStubs(ctx context.Context, projectRoot string) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Deleting all generated stubs")
	sharedFtlDir := filepath.Join(projectRoot, buildDirName)

	// Wipe the modules directory to ensure we don't have any stale modules.
	err := os.RemoveAll(sharedFtlDir)
	if err != nil {
		return fmt.Errorf("failed to remove %s: %w", sharedFtlDir, err)
	}

	return nil
}

// GenerateStubsForModules generates stubs for all modules in the schema.
func GenerateStubsForModules(ctx context.Context, projectRoot string, moduleConfigs []moduleconfig.ModuleConfig, sch *schema.Schema) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Generating module stubs")

	sharedFtlDir := filepath.Join(projectRoot, buildDirName)

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	for _, module := range sch.Modules {
		var moduleConfig *moduleconfig.ModuleConfig
		for _, mc := range moduleConfigs {
			mcCopy := mc
			if mc.Module == module.Name {
				moduleConfig = &mcCopy
				break
			}
		}

		var goModVersion string
		var replacements []*modfile.Replace
		var err error

		// If there's no module config, use the go.mod file for the first config we find.
		if moduleConfig == nil {
			if len(moduleConfigs) > 0 {
				_, goModVersion, err = updateGoModule(filepath.Join(moduleConfigs[0].Dir, "go.mod"))
				if err != nil {
					return err
				}
			} else {
				// The best we can do here if we don't have a module to read from is to use the current Go version.
				goModVersion = runtime.Version()[2:]
			}

			replacements = []*modfile.Replace{}
		} else {
			replacements, goModVersion, err = updateGoModule(filepath.Join(moduleConfig.Dir, "go.mod"))
			if err != nil {
				return err
			}
		}

		goVersion := runtime.Version()[2:]
		if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
			return fmt.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
		}

		context := ExternalModuleContext{
			Schema:       sch,
			GoVersion:    goModVersion,
			FTLVersion:   ftlVersion,
			Module:       module,
			Replacements: replacements,
		}

		funcs := maps.Clone(scaffoldFuncs)
		err = internal.ScaffoldZip(externalModuleTemplateFiles(), projectRoot, context, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs))
		if err != nil {
			return fmt.Errorf("failed to scaffold zip: %w", err)
		}

		modulesDir := filepath.Join(sharedFtlDir, "go", "modules", module.Name)
		if err := exec.Command(ctx, log.Debug, modulesDir, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
			return fmt.Errorf("failed to tidy go.mod: %w", err)
		}
	}

	return nil
}

func SyncGeneratedStubReferences(ctx context.Context, projectRootDir string, stubbedModules []string, moduleConfigs []moduleconfig.ModuleConfig) error {
	for _, moduleConfig := range moduleConfigs {
		var sharedModulesPaths []string
		for _, mod := range stubbedModules {
			if mod == moduleConfig.Module {
				continue
			}
			sharedModulesPaths = append(sharedModulesPaths, filepath.Join(projectRootDir, buildDirName, "go", "modules", mod))
		}

		_, goModVersion, err := updateGoModule(filepath.Join(moduleConfig.Dir, "go.mod"))
		if err != nil {
			return err
		}

		funcs := maps.Clone(scaffoldFuncs)
		if err := internal.ScaffoldZip(mainWorkTemplateFiles(), moduleConfig.Dir, MainWorkContext{
			GoVersion:          goModVersion,
			SharedModulesPaths: sharedModulesPaths,
		}, scaffolder.Exclude("^go.mod$"), scaffolder.Functions(funcs)); err != nil {
			return fmt.Errorf("failed to scaffold zip: %w", err)
		}
	}

	return nil
}

var scaffoldFuncs = scaffolder.FuncMap{
	"comment": schema.EncodeComments,
	"type":    genType,
	"is": func(kind string, t schema.Node) bool {
		return stdreflect.Indirect(stdreflect.ValueOf(t)).Type().Name() == kind
	},
	"imports": func(m *schema.Module) map[string]string {
		imports := map[string]string{}
		_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error { //nolint:errcheck
			switch n := n.(type) {
			case *schema.Ref:
				if n.Module == "" || n.Module == m.Name {
					break
				}
				imports[path.Join("ftl", n.Module)] = "ftl" + n.Module

				for _, tp := range n.TypeParameters {
					tpRef, err := schema.ParseRef(tp.String())
					if err != nil {
						panic(err)
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

			case *schema.TypeAlias:
				if n.IsExported() {
					if im, _ := getGoExternalTypeForWidenedType(n); im != "" {
						unquoted, err := strconv.Unquote(im)
						if err != nil {
							panic(err)
						}
						imports[unquoted] = ""
					}
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
			if i := strings.LastIndex(st.fqName, "."); i != -1 {
				lessTypeName := strings.TrimSuffix(st.fqName, st.fqName[i:])
				imports.Add(strings.TrimPrefix(lessTypeName, "ftl/"))
			}
			for _, v := range st.Variants {
				if i := strings.LastIndex(v.Type, "."); i != -1 {
					lessTypeName := strings.TrimSuffix(v.Type, v.Type[i:])
					imports.Add(strings.TrimPrefix(lessTypeName, "ftl/"))
				}
			}
		}
		out := imports.ToSlice()
		slices.Sort(out)
		return out
	},
	"typesImports": func(ctx mainModuleContext) []string {
		imports := sets.NewSet[string]()
		for _, st := range ctx.LocalSumTypes {
			if i := strings.LastIndex(st.fqName, "."); i != -1 {
				lessTypeName := strings.TrimSuffix(st.fqName, st.fqName[i:])
				// subpackage
				if len(strings.Split(lessTypeName, "/")) > 2 {
					imports.Add(lessTypeName)
				}
			}
			for _, v := range st.Variants {
				if i := strings.LastIndex(v.Type, "."); i != -1 {
					lessTypeName := strings.TrimSuffix(v.Type, v.Type[i:])
					// subpackage
					if len(strings.Split(lessTypeName, "/")) > 2 {
						imports.Add(lessTypeName)
					}
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
	"trimModuleQualifier": func(moduleName string, str string) string {
		if strings.HasPrefix(str, moduleName+".") {
			return strings.TrimPrefix(str, moduleName+".")
		}
		return str
	},
	"typeAliasType": func(m *schema.Module, t *schema.TypeAlias) string {
		if _, goType := getGoExternalTypeForWidenedType(t); goType != "" {
			return goType
		}
		return genType(m, t.Type)
	},
}

func getGoExternalTypeForWidenedType(t *schema.TypeAlias) (_import string, _type string) {
	var goType string
	var im string
	for _, md := range t.Metadata {
		md, ok := md.(*schema.MetadataTypeMap)
		if !ok {
			continue
		}

		if md.Runtime == "go" {
			var err error
			im, goType, err = getGoExternalType(md.NativeName)
			if err != nil {
				panic(err)
			}
		}
	}
	return im, goType
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
	modulepb := module.ToProto().(*schemapb.Module) //nolint:forcetypeassert
	// If user has overridden GOOS and GOARCH we want to use those values.
	goos, ok := os.LookupEnv("GOOS")
	if !ok {
		goos = runtime.GOOS
	}
	goarch, ok := os.LookupEnv("GOARCH")
	if !ok {
		goarch = runtime.GOARCH
	}

	modulepb.Runtime = &schemapb.ModuleRuntime{
		CreateTime: timestamppb.Now(),
		Language:   "go",
		Os:         &goos,
		Arch:       &goarch,
	}
	schemaBytes, err := proto.Marshal(module.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	return os.WriteFile(config.Abs().Schema, schemaBytes, 0600)
}

func writeSchemaErrors(config moduleconfig.ModuleConfig, errors []*schema.Error) error {
	el := schema.ErrorList{
		Errors: errors,
	}
	elBytes, err := proto.Marshal(el.ToProto())
	if err != nil {
		return fmt.Errorf("failed to marshal errors: %w", err)
	}
	return os.WriteFile(config.Abs().Errors, elBytes, 0600)
}

func getLocalSumTypes(module *schema.Module, nativeNames NativeNames) []goSumType {
	sumTypes := make(map[string]goSumType)
	for _, d := range module.Decls {
		e, ok := d.(*schema.Enum)
		if !ok {
			continue
		}
		if e.IsValueEnum() {
			continue
		}
		if st, ok := getGoSumType(e, nativeNames).Get(); ok {
			enumFqName := nativeNames[e]
			sumTypes[enumFqName] = st
		}
	}
	return maps.Values(sumTypes)
}

func getLocalExternalTypes(module *schema.Module) ([]goExternalType, error) {
	types := make(map[string][]string)
	for _, d := range module.Decls {
		switch d := d.(type) {
		case *schema.TypeAlias:
			var fqName string
			for _, m := range d.Metadata {
				if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
					fqName = m.NativeName
				}
			}
			if fqName == "" {
				continue
			}
			im, typ, err := getGoExternalType(fqName)
			if err != nil {
				return nil, err
			}
			if _, ok := types[im]; !ok {
				types[im] = []string{}
			}
			types[im] = append(types[im], typ)
		default:
		}
	}
	var out []goExternalType
	for im, types := range types {
		out = append(out, goExternalType{
			Import: im,
			Types:  types,
		})
	}
	return out, nil
}

// getRegisteredTypesExternalToModule returns all sum types and external types that are not defined in the given module.
// These are the types that must be registered in the main module.
func getRegisteredTypes(module *schema.Module, sch *schema.Schema, nativeNames NativeNames) ([]goSumType, []goExternalType, error) {
	sumTypes := make(map[string]goSumType)
	externalTypes := make(map[string]sets.Set[string])
	// register sum types from other modules
	for _, decl := range getRegisteredTypesExternalToModule(module, sch) {
		switch d := decl.resolved.(type) {
		case *schema.Enum:
			variants := make([]goSumTypeVariant, 0, len(d.Variants))
			for _, v := range d.Variants {
				variants = append(variants, goSumTypeVariant{ //nolint:forcetypeassert
					Name:       decl.ref.Module + "." + v.Name,
					Type:       "ftl/" + decl.ref.Module + "." + v.Name,
					SchemaType: v.Value.(*schema.TypeValue).Value,
				})
			}
			stFqName := decl.ref.Module + "." + decl.ref.Name
			sumTypes[decl.ref.ToRefKey().String()] = goSumType{
				Discriminator: stFqName,
				Variants:      variants,
			}
		case *schema.TypeAlias:
			for _, m := range d.Metadata {
				if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
					im, typ, err := getGoExternalType(m.NativeName)
					if err != nil {
						return nil, nil, err
					}
					if _, ok := externalTypes[im]; !ok {
						externalTypes[im] = sets.NewSet[string]()
					}
					externalTypes[im].Add(typ)
				}
			}
		default:
		}
	}
	for _, d := range getLocalSumTypes(module, nativeNames) {
		sumTypes[d.fqName] = d
	}
	stOut := maps.Values(sumTypes)
	slices.SortFunc(stOut, func(a, b goSumType) int {
		return strings.Compare(a.Discriminator, b.Discriminator)
	})

	localExternalTypes, err := getLocalExternalTypes(module)
	if err != nil {
		return nil, nil, err
	}
	for _, et := range localExternalTypes {
		if _, ok := externalTypes[et.Import]; !ok {
			externalTypes[et.Import] = sets.NewSet[string]()
		}
		externalTypes[et.Import].Append(et.Types...)
	}

	var etOut []goExternalType
	for im, types := range externalTypes {
		etOut = append(etOut, goExternalType{
			Import: im,
			Types:  types.ToSlice(),
		})
	}

	return stOut, etOut, nil
}

func getGoSumType(enum *schema.Enum, nativeNames NativeNames) optional.Option[goSumType] {
	if enum.IsValueEnum() {
		return optional.None[goSumType]()
	}
	variants := make([]goSumTypeVariant, 0, len(enum.Variants))
	for _, v := range enum.Variants {
		nativeName := nativeNames[v]
		lastSlash := strings.LastIndex(nativeName, "/")
		variants = append(variants, goSumTypeVariant{ //nolint:forcetypeassert
			Name:       nativeName[lastSlash+1:],
			Type:       nativeName,
			SchemaType: v.Value.(*schema.TypeValue).Value,
		})
	}
	stFqName := nativeNames[enum]
	lastSlash := strings.LastIndex(stFqName, "/")
	return optional.Some(goSumType{
		Discriminator: stFqName[lastSlash+1:],
		Variants:      variants,
		fqName:        stFqName,
	})
}

func getGoExternalType(fqName string) (_import string, _type string, err error) {
	im, err := goImportFromQualifiedName(fqName)
	if err != nil {
		return "", "", err
	}

	var pkg string
	if i := strings.LastIndex(im, " "); i != -1 {
		// import has an alias and this will be the package
		pkg = im[:i]
		im = im[i+1:]
	}
	unquoted, err := strconv.Unquote(im)
	if err != nil {
		return "", "", fmt.Errorf("failed to unquote import %q: %w", im, err)
	}
	if pkg == "" {
		pkg = unquoted[strings.LastIndex(unquoted, "/")+1:]
	}
	typeName := fqName[strings.LastIndex(fqName, ".")+1:]
	return im, fmt.Sprintf("%s.%s", pkg, typeName), nil
}

type externalDecl struct {
	ref      *schema.Ref
	resolved schema.Decl
}

// getRegisteredTypesExternalToModule returns all sum types and external types that are not defined in the given module.
// These types must be registered in the main module.
func getRegisteredTypesExternalToModule(module *schema.Module, sch *schema.Schema) []externalDecl {
	combinedSch := schema.Schema{
		Modules: append(sch.Modules, module),
	}
	var externalTypes []externalDecl
	err := schema.Visit(&combinedSch, func(n schema.Node, next func() error) error {
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		decl, ok := sch.Resolve(ref).Get()
		if !ok {
			return next()
		}
		switch d := decl.(type) {
		case *schema.Enum:
			if ref.Module != "" && ref.Module != module.Name {
				return next()
			}
			if d.IsValueEnum() {
				return next()
			}
			externalTypes = append(externalTypes, externalDecl{
				ref:      ref,
				resolved: d,
			})
		case *schema.TypeAlias:
			if len(d.Metadata) == 0 {
				return next()
			}
			externalTypes = append(externalTypes, externalDecl{
				ref:      ref,
				resolved: d,
			})
		default:
		}
		return next()
	})
	if err != nil {
		panic(fmt.Sprintf("failed to resolve external types and sum types external to the module schema: %v", err))
	}
	return externalTypes
}

// ExtractModuleSchema statically parses Go FTL module source into a schema.Module
//
// TODO: once migrated off of the legacy extractor, we can inline `extract.Extract(dir)` and delete this
// function
func ExtractModuleSchema(dir string, sch *schema.Schema) (extract.Result, error) {
	result, err := extract.Extract(dir)
	if err != nil {
		return extract.Result{}, err
	}

	// merge with legacy results for now
	if err = legacyExtractModuleSchema(dir, sch, &result); err != nil {
		return extract.Result{}, err
	}

	schema.SortErrorsByPosition(result.Errors)
	if schema.ContainsTerminalError(result.Errors) {
		return result, nil
	}
	err = schema.ValidateModule(result.Module)
	if err != nil {
		return extract.Result{}, err
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

	_ = schema.Visit(d, func(n schema.Node, next func() error) error { //nolint:errcheck
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
			updateTransitiveVisibility(decl, module)
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

// package and directory names are the same (dir=bar, pkg=bar): "github.com/foo/bar.A" => "github.com/foo/bar"
// package and directory names differ (dir=bar, pkg=baz): "github.com/foo/bar.baz.A" => "baz github.com/foo/bar"
func goImportFromQualifiedName(qualifiedName string) (string, error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return "", fmt.Errorf("invalid qualified type format %q", qualifiedName)
	}

	pkgPath := qualifiedName[:lastDotIndex]
	pkgName := path.Base(pkgPath)

	importAlias := ""
	if lastDotIndex = strings.LastIndex(pkgName, "."); lastDotIndex != -1 {
		pkgName = pkgName[lastDotIndex+1:]
		pkgPath = pkgPath[:strings.LastIndex(pkgPath, ".")]
		// package and path differ, so we need to alias the import
		importAlias = pkgName + " "
	}
	return fmt.Sprintf("%s%q", importAlias, pkgPath), nil
}
