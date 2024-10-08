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

	"github.com/TBD54566975/ftl"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/builderrors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/reflect"
	"github.com/TBD54566975/ftl/internal/schema"
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

type mainModuleContext struct {
	GoVersion          string
	FTLVersion         string
	Name               string
	SharedModulesPaths []string
	Verbs              []goVerb
	Replacements       []*modfile.Replace
	MainCtx            mainFileContext
	TypesCtx           typesFileContext
}

func (c *mainModuleContext) withImports(mainModuleImport string) {
	c.MainCtx.Imports = c.generateMainImports()
	c.TypesCtx.Imports = c.generateTypesImports(mainModuleImport)
}

func (c *mainModuleContext) generateMainImports() []string {
	imports := sets.NewSet[string]()
	imports.Add(`"context"`)
	imports.Add(`"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"`)
	imports.Add(`"github.com/TBD54566975/ftl/common/plugin"`)
	imports.Add(`"github.com/TBD54566975/ftl/go-runtime/server"`)
	if len(c.MainCtx.SumTypes) > 0 || len(c.MainCtx.ExternalTypes) > 0 {
		imports.Add(`"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"`)
	}

	for _, v := range c.Verbs {
		imports.Append(verbImports(v)...)
	}
	for _, st := range c.MainCtx.SumTypes {
		imports.Add(st.importStatement())
		for _, v := range st.Variants {
			imports.Add(v.importStatement())
		}
	}
	for _, e := range c.MainCtx.ExternalTypes {
		imports.Add(e.importStatement())
	}
	return imports.ToSlice()
}

func (c *mainModuleContext) generateTypesImports(mainModuleImport string) []string {
	imports := sets.NewSet[string]()
	if len(c.TypesCtx.SumTypes) > 0 || len(c.TypesCtx.ExternalTypes) > 0 {
		imports.Add(`"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"`)
	}
	if len(c.Verbs) > 0 {
		imports.Add(`"context"`)
	}
	for _, st := range c.TypesCtx.SumTypes {
		imports.Add(st.importStatement())
		for _, v := range st.Variants {
			imports.Add(v.importStatement())
		}
	}
	for _, et := range c.TypesCtx.ExternalTypes {
		imports.Add(et.importStatement())
	}
	for _, v := range c.Verbs {
		imports.Append(verbImports(v)...)
	}

	var filteredImports []string
	for _, im := range imports.ToSlice() {
		if im == mainModuleImport {
			continue
		}
		filteredImports = append(filteredImports, im)
	}
	return filteredImports
}

func typeImports(t goSchemaType) []string {
	imports := sets.NewSet[string]()
	if nt, ok := t.nativeType.Get(); ok {
		imports.Add(nt.importStatement())
	}
	for _, c := range t.children {
		imports.Append(typeImports(c)...)
	}
	return imports.ToSlice()
}

func verbImports(v goVerb) []string {
	imports := sets.NewSet[string]()
	imports.Add(v.importStatement())
	imports.Add(`"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"`)

	if nt, ok := v.Request.nativeType.Get(); ok && v.Request.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	if nt, ok := v.Response.nativeType.Get(); ok && v.Response.TypeName != "ftl.Unit" {
		imports.Add(nt.importStatement())
	}
	for _, r := range v.Request.children {
		imports.Append(typeImports(r)...)
	}
	for _, r := range v.Response.children {
		imports.Append(typeImports(r)...)
	}

	for _, r := range v.Resources {
		if c, ok := r.(verbClient); ok {
			imports.Add(`"github.com/TBD54566975/ftl/go-runtime/server"`)
			imports.Append(verbImports(c.goVerb)...)
		}
	}
	return imports.ToSlice()
}

type mainFileContext struct {
	Imports []string

	ProjectName   string
	SumTypes      []goSumType
	ExternalTypes []goExternalType
}

type typesFileContext struct {
	Imports       []string
	MainModulePkg string

	SumTypes      []goSumType
	ExternalTypes []goExternalType
}

type goType interface {
	getNativeType() nativeType
}

type nativeType struct {
	Name       string
	pkg        string
	importPath string
	// true if the package name differs from the directory provided by the import path
	importAlias bool
}

func (n nativeType) importStatement() string {
	if n.importAlias {
		return fmt.Sprintf("%s %q", n.pkg, n.importPath)
	}
	return strconv.Quote(n.importPath)
}

func (n nativeType) TypeName() string {
	return n.pkg + "." + n.Name
}

type goVerb struct {
	Request   goSchemaType
	Response  goSchemaType
	Resources []verbResource

	nativeType
}

type goSchemaType struct {
	TypeName      string
	LocalTypeName string
	children      []goSchemaType

	nativeType optional.Option[nativeType]
}

func (g goVerb) getNativeType() nativeType { return g.nativeType }

type goExternalType struct {
	nativeType
}

func (g goExternalType) getNativeType() nativeType { return g.nativeType }

type goSumType struct {
	Variants []goSumTypeVariant

	nativeType
}

func (g goSumType) getNativeType() nativeType { return g.nativeType }

type goSumTypeVariant struct {
	Type goSchemaType

	nativeType
}

func (g goSumTypeVariant) getNativeType() nativeType { return g.nativeType }

type verbResource interface {
	resource()
}

type verbClient struct {
	goVerb
}

func (v verbClient) resource() {}

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
func Build(ctx context.Context, projectRootDir, moduleDir string, config moduleconfig.AbsModuleConfig, sch *schema.Schema, filesTransaction ModifyFilesTransaction, buildEnv []string, devMode bool) (moduleSch *schema.Module, buildErrors []builderrors.Error, err error) {
	if err := filesTransaction.Begin(); err != nil {
		return nil, nil, fmt.Errorf("could not start a file transaction: %w", err)
	}
	defer func() {
		if terr := filesTransaction.End(); terr != nil {
			err = fmt.Errorf("failed to end file transaction: %w", terr)
		}
	}()

	replacements, goModVersion, err := updateGoModule(filepath.Join(moduleDir, "go.mod"))
	if err != nil {
		return nil, nil, err
	}

	goVersion := runtime.Version()[2:]
	if semver.Compare("v"+goVersion, "v"+goModVersion) < 0 {
		return nil, nil, fmt.Errorf("go version %q is not recent enough for this module, needs minimum version %q", goVersion, goModVersion)
	}

	ftlVersion := ""
	if ftl.IsRelease(ftl.Version) {
		ftlVersion = ftl.Version
	}

	projectName := ""
	if pcpath, ok := projectconfig.DefaultConfigPath().Get(); ok {
		pc, err := projectconfig.Load(ctx, pcpath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load project config: %w", err)
		}
		projectName = pc.Name
	}

	logger := log.FromContext(ctx)
	funcs := maps.Clone(scaffoldFuncs)

	buildDir := buildDir(moduleDir)
	err = os.MkdirAll(buildDir, 0750)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create build directory: %w", err)
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
		return nil, nil, fmt.Errorf("failed to scaffold zip: %w", err)
	}

	logger.Debugf("Extracting schema")
	result, err := extract.Extract(config.Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("could not extract schema: %w", err)
	}

	if builderrors.ContainsTerminalError(result.Errors) {
		// Only bail if schema errors contain elements at level ERROR.
		// If errors are only at levels below ERROR (e.g. INFO, WARN), the schema can still be used.
		return nil, result.Errors, nil
	}

	logger.Debugf("Generating main module")
	mctx, err := buildMainModuleContext(sch, result, goModVersion, ftlVersion, projectName, sharedModulesPaths,
		replacements)
	if err != nil {
		return nil, nil, err
	}
	if err := internal.ScaffoldZip(buildTemplateFiles(), moduleDir, mctx, scaffolder.Exclude("^go.mod$"),
		scaffolder.Functions(funcs)); err != nil {
		return nil, nil, fmt.Errorf("failed to scaffold build template: %w", err)
	}

	logger.Debugf("Tidying go.mod files")
	wg, wgctx := errgroup.WithContext(ctx)

	ftlTypesFilename := "types.ftl.go"
	wg.Go(func() error {
		if err := exec.Command(wgctx, log.Debug, moduleDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", moduleDir, err)
		}

		if err := exec.Command(wgctx, log.Debug, moduleDir, "go", "fmt", ftlTypesFilename).RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to format module dir: %w", moduleDir, err)
		}
		return filesTransaction.ModifiedFiles(filepath.Join(moduleDir, "go.mod"), filepath.Join(moduleDir, "go.sum"), filepath.Join(moduleDir, ftlTypesFilename))
	})
	mainDir := filepath.Join(buildDir, "go", "main")
	wg.Go(func() error {
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "mod", "tidy").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to tidy go.mod: %w", mainDir, err)
		}
		if err := exec.Command(wgctx, log.Debug, mainDir, "go", "fmt", "./...").RunBuffered(wgctx); err != nil {
			return fmt.Errorf("%s: failed to format main dir: %w", mainDir, err)
		}
		return filesTransaction.ModifiedFiles(filepath.Join(mainDir, "go.mod"), filepath.Join(moduleDir, "go.sum"))
	})
	if err := wg.Wait(); err != nil {
		return nil, nil, err // nolint:wrapcheck
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
		return nil, nil, fmt.Errorf("failed to compile: %w", err)
	}
	err = os.WriteFile(filepath.Join(mainDir, "../../launch"), []byte(`#!/bin/bash
	if [ -n "$FTL_DEBUG_PORT" ] && command -v dlv &> /dev/null ; then
	    dlv --listen=localhost:$FTL_DEBUG_PORT --headless=true --api-version=2 --accept-multiclient --allow-non-terminal-interactive exec --continue ./main
	else
		exec ./main
	fi
	`), 0770) // #nosec
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write launch script: %w", err)
	}
	return result.Module, result.Errors, nil
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

type mainModuleContextBuilder struct {
	sch         *schema.Schema
	mainModule  *schema.Module
	nativeNames extract.NativeNames
	imports     map[string]string
}

func buildMainModuleContext(sch *schema.Schema, result extract.Result, goModVersion, ftlVersion, projectName string,
	sharedModulesPaths []string, replacements []*modfile.Replace) (mainModuleContext, error) {
	combinedSch := &schema.Schema{
		Modules: append(sch.Modules, result.Module),
	}
	builder := &mainModuleContextBuilder{
		sch:         combinedSch,
		mainModule:  result.Module,
		nativeNames: result.NativeNames,
		imports:     imports(result.Module, false),
	}
	return builder.build(goModVersion, ftlVersion, projectName, sharedModulesPaths, replacements)
}

func (b *mainModuleContextBuilder) build(goModVersion, ftlVersion, projectName string,
	sharedModulesPaths []string, replacements []*modfile.Replace) (mainModuleContext, error) {
	ctx := &mainModuleContext{
		GoVersion:          goModVersion,
		FTLVersion:         ftlVersion,
		Name:               b.mainModule.Name,
		SharedModulesPaths: sharedModulesPaths,
		Replacements:       replacements,
		Verbs:              make([]goVerb, 0, len(b.mainModule.Decls)),
		MainCtx: mainFileContext{
			ProjectName:   projectName,
			SumTypes:      []goSumType{},
			ExternalTypes: []goExternalType{},
		},
		TypesCtx: typesFileContext{
			SumTypes:      []goSumType{},
			ExternalTypes: []goExternalType{},
		},
	}

	visited := sets.NewSet[string]()
	err := b.visit(ctx, b.mainModule, b.mainModule, visited)
	if err != nil {
		return mainModuleContext{}, err
	}

	slices.SortFunc(ctx.MainCtx.SumTypes, func(a, b goSumType) int {
		return strings.Compare(a.TypeName(), b.TypeName())
	})
	slices.SortFunc(ctx.TypesCtx.SumTypes, func(a, b goSumType) int {
		return strings.Compare(a.TypeName(), b.TypeName())
	})

	ctx.TypesCtx.MainModulePkg = b.mainModule.Name
	mainModuleImport := fmt.Sprintf("ftl/%s", b.mainModule.Name)
	if alias, ok := b.imports[mainModuleImport]; ok {
		mainModuleImport = fmt.Sprintf("%s %q", alias, mainModuleImport)
		ctx.TypesCtx.MainModulePkg = alias
	}
	ctx.withImports(mainModuleImport)
	return *ctx, nil
}

func (b *mainModuleContextBuilder) visit(
	ctx *mainModuleContext,
	module *schema.Module,
	node schema.Node,
	visited sets.Set[string],
) error {
	err := schema.Visit(node, func(node schema.Node, next func() error) error {
		if ref, ok := node.(*schema.Ref); ok {
			maybeResolved, maybeModule := b.sch.ResolveWithModule(ref)
			resolved, ok := maybeResolved.Get()
			if !ok {
				return next()
			}
			m, ok := maybeModule.Get()
			if !ok {
				return next()
			}
			err := b.visit(ctx, m, resolved, visited)
			if err != nil {
				return fmt.Errorf("failed to visit children of %s: %w", ref, err)
			}
			return next()
		}

		maybeGoType, isLocal, err := b.getGoType(module, node)
		if err != nil {
			return err
		}
		gotype, ok := maybeGoType.Get()
		if !ok {
			return next()
		}
		if visited.Contains(gotype.getNativeType().TypeName()) {
			return next()
		}
		visited.Add(gotype.getNativeType().TypeName())

		switch n := gotype.(type) {
		case goVerb:
			ctx.Verbs = append(ctx.Verbs, n)
		case goSumType:
			if isLocal {
				ctx.TypesCtx.SumTypes = append(ctx.TypesCtx.SumTypes, n)
			}
			ctx.MainCtx.SumTypes = append(ctx.MainCtx.SumTypes, n)
		case goExternalType:
			ctx.TypesCtx.ExternalTypes = append(ctx.TypesCtx.ExternalTypes, n)
			ctx.MainCtx.ExternalTypes = append(ctx.MainCtx.ExternalTypes, n)
		}
		return next()
	})
	if err != nil {
		return fmt.Errorf("failed to build main module context: %w", err)
	}
	return nil
}

func (b *mainModuleContextBuilder) getGoType(module *schema.Module, node schema.Node) (gotype optional.Option[goType], isLocal bool, err error) {
	isLocal = b.visitingMainModule(module.Name)
	switch n := node.(type) {
	case *schema.Verb:
		if !isLocal {
			return optional.None[goType](), false, nil
		}
		goverb, err := b.processVerb(n)
		if err != nil {
			return optional.None[goType](), isLocal, err
		}
		return optional.Some[goType](goverb), isLocal, nil

	case *schema.Enum:
		if n.IsValueEnum() {
			return optional.None[goType](), isLocal, nil
		}
		st, err := b.processSumType(module, n)
		if err != nil {
			return optional.None[goType](), isLocal, err
		}
		return optional.Some[goType](st), isLocal, nil

	case *schema.TypeAlias:
		if len(n.Metadata) == 0 {
			return optional.None[goType](), isLocal, nil
		}
		return b.processExternalTypeAlias(n), isLocal, nil

	default:
	}
	return optional.None[goType](), isLocal, nil
}

func (b *mainModuleContextBuilder) visitingMainModule(moduleName string) bool {
	return moduleName == b.mainModule.Name
}

func (b *mainModuleContextBuilder) processSumType(module *schema.Module, enum *schema.Enum) (goSumType, error) {
	moduleName := module.Name
	var nt nativeType
	var err error
	if !b.visitingMainModule(moduleName) {
		nt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + enum.Name)
	} else if nn, ok := b.nativeNames[enum]; ok {
		nt, err = b.getNativeType(nn)
	} else {
		return goSumType{}, fmt.Errorf("missing native name for enum %s", enum.Name)
	}
	if err != nil {
		return goSumType{}, err
	}

	variants := make([]goSumTypeVariant, 0, len(enum.Variants))
	for _, v := range enum.Variants {
		var vnt nativeType
		if !b.visitingMainModule(moduleName) {
			vnt, err = nativeTypeFromQualifiedName("ftl/" + moduleName + "." + v.Name)
		} else if nn, ok := b.nativeNames[v]; ok {
			vnt, err = b.getNativeType(nn)
		} else {
			return goSumType{}, fmt.Errorf("missing native name for enum variant %s", enum.Name)
		}
		if err != nil {
			return goSumType{}, err
		}

		typ, err := b.getGoSchemaType(v.Value.(*schema.TypeValue).Value)
		if err != nil {
			return goSumType{}, err
		}
		variants = append(variants, goSumTypeVariant{
			Type:       typ,
			nativeType: vnt,
		})
	}

	return goSumType{
		Variants:   variants,
		nativeType: nt,
	}, nil
}

func (b *mainModuleContextBuilder) processExternalTypeAlias(alias *schema.TypeAlias) optional.Option[goType] {
	for _, m := range alias.Metadata {
		if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
			nt, ok := nativeTypeForWidenedType(alias)
			if !ok {
				continue
			}
			return optional.Some[goType](goExternalType{
				nativeType: nt,
			})
		}
	}
	return optional.None[goType]()
}

func (b *mainModuleContextBuilder) processVerb(verb *schema.Verb) (goVerb, error) {
	var resources []verbResource
	for _, m := range verb.Metadata {
		switch md := m.(type) {
		case *schema.MetadataCalls:
			for _, call := range md.Calls {
				resolved, ok := b.sch.Resolve(call).Get()
				if !ok {
					return goVerb{}, fmt.Errorf("failed to resolve %s client, used by %s.%s", call,
						b.mainModule.Name, verb.Name)
				}
				callee, ok := resolved.(*schema.Verb)
				if !ok {
					return goVerb{}, fmt.Errorf("%s.%s uses %s client, but %s is not a verb",
						b.mainModule.Name, verb.Name, call, call)
				}
				calleeNativeName, ok := b.nativeNames[call]
				if !ok {
					return goVerb{}, fmt.Errorf("missing native name for verb client %s", call)
				}
				calleeverb, err := b.getGoVerb(calleeNativeName, callee)
				if err != nil {
					return goVerb{}, err
				}
				resources = append(resources, verbClient{
					calleeverb,
				})
			}
		default:
			// TODO: implement other resources
		}
	}

	nativeName, ok := b.nativeNames[verb]
	if !ok {
		return goVerb{}, fmt.Errorf("missing native name for verb %s", verb.Name)
	}
	return b.getGoVerb(nativeName, verb, resources...)
}

func (b *mainModuleContextBuilder) getGoVerb(nativeName string, verb *schema.Verb, resources ...verbResource) (goVerb, error) {
	nt, err := b.getNativeType(nativeName)
	if err != nil {
		return goVerb{}, err
	}
	req, err := b.getGoSchemaType(verb.Request)
	if err != nil {
		return goVerb{}, err
	}
	resp, err := b.getGoSchemaType(verb.Response)
	if err != nil {
		return goVerb{}, err
	}
	return goVerb{
		nativeType: nt,
		Request:    req,
		Response:   resp,
		Resources:  resources,
	}, nil
}

func (b *mainModuleContextBuilder) getGoSchemaType(typ schema.Type) (goSchemaType, error) {
	result := goSchemaType{
		TypeName:      genTypeWithNativeNames(nil, typ, b.nativeNames),
		LocalTypeName: genTypeWithNativeNames(b.mainModule, typ, b.nativeNames),
		children:      []goSchemaType{},
		nativeType:    optional.None[nativeType](),
	}

	nn, ok := b.nativeNames[typ]
	if ok {
		nt, err := b.getNativeType(nn)
		if err != nil {
			return goSchemaType{}, err
		}
		result.nativeType = optional.Some(nt)
	}

	switch t := typ.(type) {
	case *schema.Ref:
		// we add native types for all refs traversed from the main module. however, if this
		// ref was not directly traversed by the main module (e.g. the request of a verb in an external module, where
		// the main module needs a generated client for this verb), we can infer its native qualified name to get the
		// native type here.
		if !result.nativeType.Ok() {
			nt, err := b.getNativeType("ftl/" + t.Module + "." + t.Name)
			if err != nil {
				return goSchemaType{}, err
			}
			result.nativeType = optional.Some(nt)
		}
		if len(t.TypeParameters) > 0 {
			for _, tp := range t.TypeParameters {
				_r, err := b.getGoSchemaType(tp)
				if err != nil {
					return goSchemaType{}, err
				}
				result.children = append(result.children, _r)
			}
		}
	case *schema.Time:
		nt, err := b.getNativeType("time.Time")
		if err != nil {
			return goSchemaType{}, err
		}
		result.nativeType = optional.Some(nt)
	default:
	}

	return result, nil
}

func (b *mainModuleContextBuilder) getNativeType(qualifiedName string) (nativeType, error) {
	nt, err := nativeTypeFromQualifiedName(qualifiedName)
	if err != nil {
		return nativeType{}, err
	}
	// we already have an alias name for this import path
	if alias, ok := b.imports[nt.importPath]; ok {
		if alias != path.Base(nt.importPath) {
			nt.pkg = alias
			nt.importAlias = true
		}
		return nt, nil
	}
	b.imports = addImports(b.imports, nt)
	return nt, nil
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
		for _, md := range t.Metadata {
			md, ok := md.(*schema.MetadataTypeMap)
			if !ok || md.Runtime != "go" {
				continue
			}
			nt, err := nativeTypeFromQualifiedName(md.NativeName)
			if err != nil {
				return ""
			}
			return fmt.Sprintf("%s.%s", nt.pkg, nt.Name)
		}
		return genType(m, t.Type)
	},
	"getVerbClient": func(resource verbResource) *verbClient {
		if c, ok := resource.(verbClient); ok {
			return &c
		}
		return nil
	},
}

// returns the import path and the directory name for a type alias if there is an associated go library
func nativeTypeForWidenedType(t *schema.TypeAlias) (nt nativeType, ok bool) {
	for _, md := range t.Metadata {
		md, ok := md.(*schema.MetadataTypeMap)
		if !ok {
			continue
		}

		if md.Runtime == "go" {
			var err error
			goType, err := nativeTypeFromQualifiedName(md.NativeName)
			if err != nil {
				panic(err)
			}
			return goType, true
		}
	}
	return nt, false
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
	return genTypeWithNativeNames(module, t, nil)
}

// TODO: this is a hack because we don't currently qualify schema refs. Using native names for now to ensure
// even if the module is the same, we qualify the type with a package name when it's a subpackage.
func genTypeWithNativeNames(module *schema.Module, t schema.Type, nativeNames extract.NativeNames) string {
	switch t := t.(type) {
	case *schema.Ref:
		pkg := "ftl" + t.Module
		name := t.Name
		if nativeNames != nil {
			if nn, ok := nativeNames[t]; ok {
				nt, err := nativeTypeFromQualifiedName(nn)
				if err == nil {
					pkg = nt.pkg
					name = nt.Name
				}
			}
		}

		desc := ""
		if module != nil && pkg == "ftl"+module.Name {
			desc = name
		} else if t.Module == "" {
			desc = name
		} else {
			desc = pkg + "." + name
		}
		if len(t.TypeParameters) > 0 {
			desc += "["
			for i, tp := range t.TypeParameters {
				if i != 0 {
					desc += ", "
				}
				desc += genTypeWithNativeNames(module, tp, nativeNames)
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
		return "[]" + genTypeWithNativeNames(module, t.Element, nativeNames)

	case *schema.Map:
		return "map[" + genTypeWithNativeNames(module, t.Key, nativeNames) + "]" + genType(module, t.Value)

	case *schema.Optional:
		return "ftl.Option[" + genTypeWithNativeNames(module, t.Type, nativeNames) + "]"

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

// returns the import path and directory name for a Go type
// package and directory names are the same (dir=bar, pkg=bar): "github.com/foo/bar.A" => "github.com/foo/bar", none
// package and directory names differ (dir=bar, pkg=baz): "github.com/foo/bar.baz.A" => "github.com/foo/bar", "baz"
func nativeTypeFromQualifiedName(qualifiedName string) (nativeType, error) {
	lastDotIndex := strings.LastIndex(qualifiedName, ".")
	if lastDotIndex == -1 {
		return nativeType{}, fmt.Errorf("invalid qualified type format %q", qualifiedName)
	}

	pkgPath := qualifiedName[:lastDotIndex]
	typeName := qualifiedName[lastDotIndex+1:]
	pkgName := path.Base(pkgPath)
	aliased := false

	if strings.LastIndex(pkgName, ".") != -1 {
		lastDotIndex = strings.LastIndex(pkgPath, ".")
		pkgName = pkgPath[lastDotIndex+1:]
		pkgPath = pkgPath[:lastDotIndex]
		aliased = true
	}

	if parts := strings.Split(qualifiedName, "/"); len(parts) > 0 && parts[0] == "ftl" {
		aliased = true
		pkgName = "ftl" + pkgName
	}

	return nativeType{
		Name:        typeName,
		pkg:         pkgName,
		importPath:  pkgPath,
		importAlias: aliased,
	}, nil
}

// imports returns a map of import paths to aliases for a module.
// - hardcoded for time ("stdtime")
// - prefixed with "ftl" for other modules (eg "ftlfoo")
// - addImports() is used to generate shortest unique aliases for external packages
func imports(m *schema.Module, aliasesMustBeExported bool) map[string]string {
	// find all imports
	imports := map[string]string{}
	// map from import path to the first dir we see
	extraImports := map[string]nativeType{}
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
			if nt, ok := nativeTypeForWidenedType(n); ok {
				if existing, ok := extraImports[nt.importPath]; !ok || !existing.importAlias {
					extraImports[nt.importPath] = nt
				}
			}
		default:
		}
		return next()
	})

	return addImports(imports, maps.Values(extraImports)...)
}

// addImports takes existing imports (mapping import path to pkg alias) and adds new imports by generating aliases
// aliases are generated for external types by finding the shortest unique alias that can be used without conflict:
func addImports(existingImports map[string]string, newTypes ...nativeType) map[string]string {
	imports := maps.Clone(existingImports)
	// maps import path to possible aliases, shortest to longest
	aliasesForImports := map[string][]string{}

	// maps possible aliases with the count of imports that could use the alias
	possibleImportAliases := map[string]int{}
	for _, alias := range imports {
		possibleImportAliases[alias]++
	}
	for _, nt := range newTypes {
		if _, ok := imports[nt.importPath]; ok {
			continue
		}

		importPath := nt.importPath
		pathComponents := strings.Split(importPath, "/")
		if nt.importAlias {
			pathComponents = append(pathComponents, nt.pkg)
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
