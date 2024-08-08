package schema

import (
	"fmt"
	"go/types"
	"slices"
	"strings"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	checker "github.com/TBD54566975/golang-tools/go/analysis/programmaticchecker"
	"github.com/TBD54566975/golang-tools/go/packages"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/tuple"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/call"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/go-runtime/schema/configsecret"
	"github.com/TBD54566975/ftl/go-runtime/schema/data"
	"github.com/TBD54566975/ftl/go-runtime/schema/database"
	"github.com/TBD54566975/ftl/go-runtime/schema/enum"
	"github.com/TBD54566975/ftl/go-runtime/schema/finalize"
	"github.com/TBD54566975/ftl/go-runtime/schema/fsm"
	"github.com/TBD54566975/ftl/go-runtime/schema/initialize"
	"github.com/TBD54566975/ftl/go-runtime/schema/metadata"
	"github.com/TBD54566975/ftl/go-runtime/schema/subscription"
	"github.com/TBD54566975/ftl/go-runtime/schema/topic"
	"github.com/TBD54566975/ftl/go-runtime/schema/transitive"
	"github.com/TBD54566975/ftl/go-runtime/schema/typealias"
	"github.com/TBD54566975/ftl/go-runtime/schema/typeenum"
	"github.com/TBD54566975/ftl/go-runtime/schema/typeenumvariant"
	"github.com/TBD54566975/ftl/go-runtime/schema/valueenumvariant"
	"github.com/TBD54566975/ftl/go-runtime/schema/verb"
)

// Extractors contains all schema extractors that will run.
//
// It is a list of lists, where each list is a round of tasks dependent on the prior round's execution (e.g. an analyzer
// in Extractors[1] will only execute once all analyzers in Extractors[0] complete). Elements of the same list
// should be considered unordered and may run in parallel.
var Extractors [][]*analysis.Analyzer

func init() {
	Extractors = [][]*analysis.Analyzer{
		{
			initialize.Analyzer,
			inspect.Analyzer,
		},
		{
			metadata.Extractor,
		},
		{
			// must run before typeenumvariant.Extractor; typeenum.Extractor determines all possible discriminator
			// interfaces and typeenumvariant.Extractor determines any types that implement these
			typeenum.Extractor,
		},
		{
			configsecret.Extractor,
			data.Extractor,
			database.Extractor,
			fsm.Extractor,
			topic.Extractor,
			typealias.Extractor,
			typeenumvariant.Extractor,
			valueenumvariant.Extractor,
			verb.Extractor,
		},
		{
			call.Extractor,
			// must run after valueenumvariant.Extractor and typeenumvariant.Extractor;
			// visits a node and aggregates its enum variants if present
			enum.Extractor,
			subscription.Extractor,
		},
		{
			transitive.Extractor,
		},
		{
			finalize.Analyzer,
		},
	}
}

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Node]string

// Result contains the final schema extraction result.
type Result struct {
	// Module is the extracted module schema.
	Module *schema.Module
	// NativeNames maps schema nodes to their native Go names.
	NativeNames NativeNames
	// Errors is a list of errors encountered during schema extraction.
	Errors []*schema.Error
}

// Extract statically parses Go FTL module source into a schema.Module
func Extract(moduleDir string) (Result, error) {
	pkgConfig := packages.Config{
		Dir:  moduleDir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}
	cConfig := checker.Config{
		LoadConfig:                  pkgConfig,
		ReverseImportExecutionOrder: true,
		Patterns:                    []string{"./..."},
	}
	results, diagnostics, err := checker.Run(cConfig, analyzersWithDependencies()...)
	if err != nil {
		return Result{}, err
	}
	return combineAllPackageResults(results, diagnostics)
}

type refResultType int

const (
	failed refResultType = iota
	widened
)

type refResult struct {
	typ    refResultType
	obj    types.Object
	fqName optional.Option[string]
}

type combinedData struct {
	module *schema.Module
	errs   []*schema.Error

	nativeNames         NativeNames
	functionCalls       map[types.Object]sets.Set[types.Object]
	verbCalls           map[types.Object]sets.Set[*schema.Ref]
	refResults          map[schema.RefKey]refResult
	extractedDecls      map[schema.Decl]types.Object
	externalTypeAliases sets.Set[*schema.TypeAlias]
	// for detecting duplicates
	typeUniqueness   map[string]tuple.Pair[types.Object, schema.Position]
	globalUniqueness map[string]tuple.Pair[types.Object, schema.Position]
}

func newCombinedData(diagnostics []analysis.SimpleDiagnostic) *combinedData {
	return &combinedData{
		errs:                diagnosticsToSchemaErrors(diagnostics),
		nativeNames:         make(NativeNames),
		functionCalls:       make(map[types.Object]sets.Set[types.Object]),
		verbCalls:           make(map[types.Object]sets.Set[*schema.Ref]),
		refResults:          make(map[schema.RefKey]refResult),
		extractedDecls:      make(map[schema.Decl]types.Object),
		externalTypeAliases: sets.NewSet[*schema.TypeAlias](),
		typeUniqueness:      make(map[string]tuple.Pair[types.Object, schema.Position]),
		globalUniqueness:    make(map[string]tuple.Pair[types.Object, schema.Position]),
	}
}

func (cd *combinedData) error(err *schema.Error) {
	cd.errs = append(cd.errs, err)
}

func (cd *combinedData) update(fr finalize.Result) {
	for decl, obj := range fr.Extracted {
		cd.validateDecl(decl, obj)
		cd.extractedDecls[decl] = obj
	}
	copyFailedRefs(cd.refResults, fr.Failed)
	maps.Copy(cd.nativeNames, fr.NativeNames)
	maps.Copy(cd.functionCalls, fr.FunctionCalls)
	maps.Copy(cd.verbCalls, fr.VerbCalls)
}

func (cd *combinedData) toResult() Result {
	cd.module.AddDecls(maps.Keys(cd.extractedDecls))
	cd.updateDeclVisibility()
	cd.propagateTypeErrors()
	schema.SortErrorsByPosition(cd.errs)
	return Result{
		Module:      cd.module,
		NativeNames: cd.nativeNames,
		Errors:      cd.errs,
	}
}

func (cd *combinedData) updateModule(fr finalize.Result) error {
	if cd.module == nil {
		cd.module = &schema.Module{Name: fr.ModuleName, Comments: fr.ModuleComments}
	} else {
		if cd.module.Name != fr.ModuleName {
			return fmt.Errorf("unexpected schema extraction result module name: %s", fr.ModuleName)
		}
		if len(cd.module.Comments) == 0 {
			cd.module.Comments = fr.ModuleComments
		}
	}
	return nil
}

func (cd *combinedData) validateDecl(decl schema.Decl, obj types.Object) {
	typename := common.GetDeclTypeName(decl)
	typeKey := fmt.Sprintf("%s-%s", typename, decl.GetName())
	if value, ok := cd.typeUniqueness[typeKey]; ok && value.A != obj {
		cd.error(schema.Errorf(decl.Position(), decl.Position().Column,
			"duplicate %s declaration for %q; already declared at %q", typename,
			cd.module.Name+"."+decl.GetName(), value.B))
	} else if value, ok := cd.globalUniqueness[decl.GetName()]; ok && value.A != obj {
		cd.error(schema.Errorf(decl.Position(), decl.Position().Column,
			"schema declaration with name %q already exists for module %q; previously declared at %q",
			decl.GetName(), cd.module.Name, value.B))
	}
	cd.typeUniqueness[typeKey] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
	cd.globalUniqueness[decl.GetName()] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
}

func (cd *combinedData) getVerbCalls(obj types.Object) sets.Set[*schema.Ref] {
	calls := sets.NewSet[*schema.Ref]()
	if cls, ok := cd.verbCalls[obj]; ok {
		calls.Append(cls.ToSlice()...)
	}
	if fnCall, ok := cd.functionCalls[obj]; ok {
		for _, calleeObj := range fnCall.ToSlice() {
			calls.Append(cd.getVerbCalls(calleeObj).ToSlice()...)
		}
	}
	return calls
}

// updateDeclVisibility traverses the module schema via refs and updates visibility as needed.
func (cd *combinedData) updateDeclVisibility() {
	for _, d := range cd.module.Decls {
		if d.IsExported() {
			updateTransitiveVisibility(d, cd.module)
		}
	}
}

// propagateTypeErrors propagates type errors to referencing nodes. This improves error messaging for the LSP client by
// surfacing errors all the way up the schema chain.
func (cd *combinedData) propagateTypeErrors() {
	_ = schema.VisitWithParent(cd.module, nil, func(n schema.Node, p schema.Node, next func() error) error { //nolint:errcheck
		if p == nil {
			return next()
		}
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		result, ok := cd.refResults[ref.ToRefKey()]
		if !ok {
			return next()
		}

		switch result.typ {
		case failed:
			refNativeName := common.GetNativeName(result.obj)
			switch pt := p.(type) {
			case *schema.Verb:
				if pt.Request == n {
					cd.error(schema.Errorf(pt.Request.Position(), pt.Request.Position().Column,
						"unsupported request type %q", refNativeName))
				}
				if pt.Response == n {
					cd.error(schema.Errorf(pt.Response.Position(), pt.Response.Position().Column,
						"unsupported response type %q", refNativeName))
				}
			case *schema.Field:
				cd.error(schema.Errorf(pt.Position(), pt.Position().Column, "unsupported type %q for "+
					"field %q", refNativeName, pt.Name))
			default:
				cd.error(schema.Errorf(p.Position(), p.Position().Column, "unsupported type %q",
					refNativeName))
			}
		case widened:
			cd.error(schema.Warnf(n.Position(), n.Position().Column, "external type %q will be "+
				"widened to Any", result.fqName.MustGet()))
		}

		return next()
	})
}

func analyzersWithDependencies() []*analysis.Analyzer {
	var as []*analysis.Analyzer
	// observes dependencies as specified by tiered list ordering in Extractors and applies the dependency
	// requirements to the analyzers
	//
	// flattens Extractors (a list of lists) into a single list to provide as input for the checker
	for i, extractorRound := range Extractors {
		for _, extractor := range extractorRound {
			extractor.RunDespiteErrors = true
			extractor.Requires = append(extractor.Requires, dependenciesBeforeIndex(i)...)
			as = append(as, extractor)
		}
	}
	return as
}

func dependenciesBeforeIndex(idx int) []*analysis.Analyzer {
	var deps []*analysis.Analyzer
	for i := range idx {
		for _, extractor := range Extractors[i] {
			if extractor == nil {
				panic(fmt.Sprintf("analyzer at Extractors[%d] not yet initialized", i))
			}

			deps = append(deps, extractor)
		}
	}
	return deps
}

func combineAllPackageResults(results map[*analysis.Analyzer][]any, diagnostics []analysis.SimpleDiagnostic) (Result, error) {
	cd := newCombinedData(diagnostics)

	fResults, ok := results[finalize.Analyzer]
	if !ok {
		return Result{}, fmt.Errorf("schema extraction finalizer result not found")
	}
	for _, r := range fResults {
		fr, ok := r.(finalize.Result)
		if !ok {
			return Result{}, fmt.Errorf("unexpected schema extraction result type: %T", r)
		}
		if err := cd.updateModule(fr); err != nil {
			return Result{}, err
		}
		cd.update(fr)
	}

	for decl, obj := range cd.extractedDecls {
		moduleName := cd.module.Name
		switch d := decl.(type) {
		case *schema.TypeAlias:
			if len(d.Metadata) > 0 {
				fqName, err := goQualifiedNameForWidenedType(obj, d.Metadata)
				if err != nil {
					cd.error(&schema.Error{Pos: d.Position(), EndColumn: d.Pos.Column,
						Msg: err.Error(), Level: schema.ERROR})
				}
				cd.refResults[schema.RefKey{Module: moduleName, Name: d.Name}] = refResult{typ: widened, obj: obj,
					fqName: optional.Some(fqName)}
				cd.externalTypeAliases.Add(d)
				cd.nativeNames[d] = common.GetNativeName(obj)
			}
		case *schema.Verb:
			calls := cd.getVerbCalls(obj).ToSlice()
			slices.SortFunc(calls, func(i, j *schema.Ref) int {
				if i.Module != j.Module {
					return strings.Compare(i.Module, j.Module)
				}
				return strings.Compare(i.Name, j.Name)
			})
			if len(calls) > 0 {
				d.Metadata = append(d.Metadata, &schema.MetadataCalls{Calls: calls})
			}
		default:
		}
	}

	result := cd.toResult()
	if schema.ContainsTerminalError(result.Errors) {
		return result, nil
	}
	return result, schema.ValidateModule(result.Module) //nolint:wrapcheck
}

// updateTransitiveVisibility updates any decls that are transitively visible from d.
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

func diagnosticsToSchemaErrors(diagnostics []analysis.SimpleDiagnostic) []*schema.Error {
	if len(diagnostics) == 0 {
		return nil
	}
	errors := make([]*schema.Error, 0, len(diagnostics))
	for _, d := range diagnostics {
		errors = append(errors, &schema.Error{
			Pos:       simplePosToSchemaPos(d.Pos),
			EndColumn: d.End.Column,
			Msg:       d.Message,
			Level:     common.DiagnosticCategory(d.Category).ToErrorLevel(),
		})
	}
	return errors
}

func copyFailedRefs(parsedRefs map[schema.RefKey]refResult, failedRefs map[schema.RefKey]types.Object) {
	for ref, obj := range failedRefs {
		parsedRefs[ref] = refResult{typ: failed, obj: obj}
	}
}

func goQualifiedNameForWidenedType(obj types.Object, metadata []schema.Metadata) (string, error) {
	var nativeName string
	for _, m := range metadata {
		if m, ok := m.(*schema.MetadataTypeMap); ok && m.Runtime == "go" {
			if nativeName != "" {
				return "", fmt.Errorf("multiple Go type mappings found for %q", common.GetNativeName(obj))
			}
			nativeName = m.NativeName
		}
	}
	if len(metadata) > 0 && nativeName == "" {
		return "", fmt.Errorf("missing Go native name in typemapped alias for %q",
			common.GetNativeName(obj))
	}
	return nativeName, nil
}

func simplePosToSchemaPos(pos analysis.SimplePosition) schema.Position {
	return schema.Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}
