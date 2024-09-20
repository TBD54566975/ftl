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

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
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
	ProjectName        string
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
func Build(ctx context.Context, projectRootDir, moduleDir string, sch *schema.Schema, filesTransaction ModifyFilesTransaction, buildEnv []string, devMode bool) (err error) {
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

	projectName := ""
	if pcpath, ok := projectconfig.DefaultConfigPath().Get(); ok {
		pc, err := projectconfig.Load(ctx, pcpath)
		if err != nil {
			return fmt.Errorf("failed to load project config: %w", err)
		}
		projectName = pc.Name
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
	result, err := extract.Extract(config.Dir)
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
		ProjectName:        projectName,
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
	args := []string{"build", "-o", "../../main", "."}
	if devMode {
		args = []string{"build", "-gcflags=all=-N -l", "-o", "../../main", "."}
	}
	// We have seen lots of upstream HTTP/2 failures that make CI unstable.
	// Disable HTTP/2 for now during the build. This can probably be removed later
	buildEnv = append(buildEnv, "GODEBUG=http2client=0")
	err = exec.CommandWithEnv(ctx, log.Debug, mainDir, buildEnv, "go", args...).RunBuffered(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile: %w", err)
	}
	err = os.WriteFile(filepath.Join(mainDir, "../../launch"), []byte(`#!/bin/bash
	if [ -n "$FTL_DEBUG_PORT" ] && command -v dlv &> /dev/null ; then
	    dlv --listen=localhost:$FTL_DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --allow-non-terminal-interactive exec --continue ./main
	else
		exec ./main
	fi
	`), 0770) // #nosec
	if err != nil {
		return fmt.Errorf("failed to write launch script: %w", err)
	}
	return nil
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
	logger.Debugf("Generating go module stubs")

	sharedFtlDir := filepath.Join(projectRoot, buildDirName)

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}
	hasGo := false
	for _, mc := range moduleConfigs {
		if mc.Language == "go" && mc.Module != "builtin" {
			hasGo = true
		}
	}
	if !hasGo {
		return nil
	}

	for _, module := range sch.Modules {
		var moduleConfig *moduleconfig.ModuleConfig
		for _, mc := range moduleConfigs {
			mcCopy := mc
			if mc.Module == module.Name && mc.Language == "go" {
				moduleConfig = &mcCopy
				break
			}
		}

		var goModVersion string
		var replacements []*modfile.Replace
		var err error

		// If there's no module config, use the go.mod file for the first config we find.
		if moduleConfig == nil {
			for _, mod := range moduleConfigs {
				if mod.Language != "go" {
					continue
				}
				goModPath := filepath.Join(mod.Dir, "go.mod")
				_, goModVersion, err = updateGoModule(goModPath)
				if err != nil {
					logger.Debugf("could not read go.mod %s", goModPath)
					continue
				}
			}
			if goModVersion == "" {
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
		if moduleConfig.Language != "go" {
			continue
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
		return imports(m, true)
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
		// it is wasteful to calculate imports each time
		imports := imports(m, false)

		for _, md := range t.Metadata {
			md, ok := md.(*schema.MetadataTypeMap)
			if !ok || md.Runtime != "go" {
				continue
			}
			if goType, err := getGoExternalType(imports, md.NativeName); err == nil {
				return goType
			}
		}
		return genType(m, t.Type)
	},
}

// returns the import path and the directory name for a type alias if there is an associated go library
func goImportForWidenedType(t *schema.TypeAlias) (importPath string, dirName optional.Option[string], ok bool) {
	for _, md := range t.Metadata {
		md, ok := md.(*schema.MetadataTypeMap)
		if !ok {
			continue
		}

		if md.Runtime == "go" {
			var err error
			importPath, dirName, err = goImportFromQualifiedName(md.NativeName)
			if err != nil {
				panic(err)
			}
			return importPath, dirName, true
		}
	}
	return importPath, dirName, false
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

func getLocalSumTypes(module *schema.Module, nativeNames extract.NativeNames) []goSumType {
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
	imports := imports(module, false)
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
			im, _, ok := goImportForWidenedType(d)
			if !ok {
				continue
			}
			typ, err := getGoExternalType(imports, fqName)
			if err != nil {
				return nil, err
			}

			importStatement := strconv.Quote(im)
			if imports[im] != "" {
				importStatement = imports[im] + " " + importStatement
			}
			if _, ok := types[importStatement]; !ok {
				types[importStatement] = []string{}
			}
			types[importStatement] = append(types[importStatement], typ)
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
func getRegisteredTypes(module *schema.Module, sch *schema.Schema, nativeNames extract.NativeNames) ([]goSumType, []goExternalType, error) {
	sumTypes := make(map[string]goSumType)
	externalTypes := make(map[string]sets.Set[string])
	externalDecls := getRegisteredTypesExternalToModule(module, sch)

	// calculate all imports and aliases
	imports := imports(module, false)
	extraImports := map[string]optional.Option[string]{}
	for _, d := range externalDecls {
		d, ok := d.resolved.(*schema.TypeAlias)
		if !ok {
			continue
		}
		for _, m := range d.Metadata {
			m, ok := m.(*schema.MetadataTypeMap)
			if !ok || m.Runtime != "go" {
				continue
			}
			if im, dirName, ok := goImportForWidenedType(d); ok && extraImports[im] == optional.None[string]() {
				extraImports[im] = dirName
			}
		}
	}
	imports = addImports(imports, extraImports)

	// register sum types from other modules
	for _, decl := range externalDecls {
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
					im, _, ok := goImportForWidenedType(d)
					if !ok {
						continue
					}
					typ, err := getGoExternalType(imports, m.NativeName)
					if err != nil {
						return nil, nil, err
					}
					importStatement := strconv.Quote(im)
					if imports[im] != "" {
						importStatement = imports[im] + " " + importStatement
					}
					if _, ok := externalTypes[importStatement]; !ok {
						externalTypes[importStatement] = sets.NewSet[string]()
					}
					externalTypes[importStatement].Add(typ)
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

func getGoSumType(enum *schema.Enum, nativeNames extract.NativeNames) optional.Option[goSumType] {
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

func getGoExternalType(imports map[string]string, fqName string) (string, error) {
	im, _, err := goImportFromQualifiedName(fqName)
	if err != nil {
		return "", err
	}
	pkg := imports[im]
	if pkg == "" {
		pkg = im[strings.LastIndex(im, "/")+1:]
	}
	typeName := fqName[strings.LastIndex(fqName, ".")+1:]
	return fmt.Sprintf("%s.%s", pkg, typeName), nil
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

// returns the import path and directory name for an external type
// package and directory names are the same (dir=bar, pkg=bar): "github.com/foo/bar.A" => "github.com/foo/bar", none
// package and directory names differ (dir=bar, pkg=baz): "github.com/foo/bar.baz.A" => "github.com/foo/bar", "baz"
func goImportFromQualifiedName(qualifiedName string) (importPath string, directoryName optional.Option[string], err error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return "", optional.None[string](), fmt.Errorf("invalid qualified type format %q", qualifiedName)
	}

	pkgPath := qualifiedName[:lastDotIndex]
	pkgName := path.Base(pkgPath)

	if lastDotIndex = strings.LastIndex(pkgName, "."); lastDotIndex != -1 {
		pkgName = pkgName[lastDotIndex+1:]
		pkgPath = pkgPath[:strings.LastIndex(pkgPath, ".")]
		return pkgPath, optional.Some(pkgName), nil
	}
	return pkgPath, optional.None[string](), nil
}

// imports returns a map of import paths to aliases for a module.
// - hardcoded for time ("stdtime")
// - prefixed with "ftl" for other modules (eg "ftlfoo")
// - addImports() is used to generate shortest unique aliases for external packages
func imports(m *schema.Module, aliasesMustBeExported bool) map[string]string {
	// find all imports
	imports := map[string]string{}
	// map from import path to the first dir we see
	extraImports := map[string]optional.Option[string]{}
	_ = schema.VisitExcludingMetadataChildren(m, func(n schema.Node, next func() error) error { //nolint:errcheck
		switch n := n.(type) {
		case *schema.Ref:
			if n.Module == "" || n.Module == m.Name {
				break
			}
			imports[path.Join("ftl", n.Module)] = "ftl" + n.Module
			for _, tp := range n.TypeParameters {
				if tpRef, ok := tp.(*schema.Ref); ok && tpRef.Module != "" && tpRef.Module != m.Name {
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
			if aliasesMustBeExported && !n.IsExported() {
				return next()
			}
			if importPath, dirName, ok := goImportForWidenedType(n); ok && extraImports[importPath] == optional.None[string]() {
				extraImports[importPath] = dirName
			}
		default:
		}
		return next()
	})

	return addImports(imports, extraImports)
}

// addImports takes existing imports (mapping import path to pkg alias) and adds new imports by generating aliases
// aliases are generated for external types by finding the shortest unique alias that can be used without conflict:
func addImports(existingImports map[string]string, newImportPathsAndDirs map[string]optional.Option[string]) map[string]string {
	imports := maps.Clone(existingImports)
	// maps import path to possible aliases, shortest to longest
	aliasesForImports := map[string][]string{}

	// maps possible aliases with the count of imports that could use the alias
	possibleImportAliases := map[string]int{}
	for _, alias := range imports {
		possibleImportAliases[alias]++
	}
	for importPath, dirName := range newImportPathsAndDirs {
		pathComponents := strings.Split(importPath, "/")
		if dirName, ok := dirName.Get(); ok {
			pathComponents = append(pathComponents, dirName)
		}

		var currentAlias string
		for i := range len(pathComponents) {
			runes := []rune(pathComponents[len(pathComponents)-1-i])
			for i, char := range runes {
				if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
					runes[i] = '_'
				}
			}
			if unicode.IsNumber(runes[0]) {
				newRunes := make([]rune, len(runes)+1)
				newRunes[0] = '_'
				copy(newRunes[1:], runes)
				runes = newRunes
			}
			foldedComponent := string(runes)
			if i == 0 {
				currentAlias = foldedComponent
			} else {
				currentAlias = foldedComponent + "_" + currentAlias
			}
			aliasesForImports[importPath] = append(aliasesForImports[importPath], currentAlias)
			possibleImportAliases[currentAlias]++
		}
	}
	for importPath, aliases := range aliasesForImports {
		found := false
		for _, alias := range aliases {
			if possibleImportAliases[alias] == 1 {
				imports[importPath] = alias
				found = true
				break
			}
		}
		if !found {
			// no possible alias that is unique, use the last one as no other type will choose the same
			imports[importPath] = aliases[len(aliases)-1]
		}
	}
	return imports
}
