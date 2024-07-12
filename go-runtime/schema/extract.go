package schema

import (
	"fmt"
	"go/types"

	"github.com/TBD54566975/ftl/go-runtime/schema/call"
	"github.com/TBD54566975/ftl/go-runtime/schema/configsecret"
	"github.com/TBD54566975/ftl/go-runtime/schema/data"
	"github.com/TBD54566975/ftl/go-runtime/schema/enum"
	"github.com/TBD54566975/ftl/go-runtime/schema/subscription"
	"github.com/TBD54566975/ftl/go-runtime/schema/topic"
	"github.com/TBD54566975/ftl/go-runtime/schema/typeenum"
	"github.com/TBD54566975/ftl/go-runtime/schema/typeenumvariant"
	"github.com/TBD54566975/ftl/go-runtime/schema/valueenumvariant"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/tuple"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/go-runtime/schema/finalize"
	"github.com/TBD54566975/ftl/go-runtime/schema/initialize"
	"github.com/TBD54566975/ftl/go-runtime/schema/metadata"
	"github.com/TBD54566975/ftl/go-runtime/schema/transitive"
	"github.com/TBD54566975/ftl/go-runtime/schema/typealias"
	"github.com/TBD54566975/ftl/go-runtime/schema/verb"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	checker "github.com/TBD54566975/golang-tools/go/analysis/programmaticchecker"
	"github.com/TBD54566975/golang-tools/go/packages"
)

// Extractors contains all schema extractors that will run.
//
// It is a list of lists, where each list is a round of tasks dependent on the prior round's execution (e.g. an analyzer
// in Extractors[1] will only execute once all analyzers in Extractors[0] complete). Elements of the same list
// should be considered unordered and may run in parallel.
var Extractors = [][]*analysis.Analyzer{
	{
		initialize.Analyzer,
		inspect.Analyzer,
	},
	{
		call.Extractor,
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
		topic.Extractor,
		typealias.Extractor,
		typeenumvariant.Extractor,
		valueenumvariant.Extractor,
		verb.Extractor,
	},
	{
		// must run after valueenumvariant.Extractor and typeenumvariant.Extractor;
		// visits a node and aggregates its enum variants if present
		enum.Extractor,
		// must run after topic.Extractor
		subscription.Extractor,
	},
	{
		transitive.Extractor,
	},
	{
		finalize.Analyzer,
	},
}

// Result contains the final schema extraction result.
type Result struct {
	// Module is the extracted module schema.
	Module *schema.Module
	// NativeNames maps schema nodes to their native Go names.
	NativeNames map[schema.Node]string
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

// the run will produce finalizer results for all packages it executes on, so we need to aggregate the results into a
// single schema
func combineAllPackageResults(results map[*analysis.Analyzer][]any, diagnostics []analysis.SimpleDiagnostic) (Result, error) {
	fResults, ok := results[finalize.Analyzer]
	if !ok {
		return Result{}, fmt.Errorf("schema extraction finalizer result not found")
	}
	combined := Result{
		NativeNames: make(map[schema.Node]string),
		Errors:      diagnosticsToSchemaErrors(diagnostics),
	}
	refResults := make(map[schema.RefKey]refResult)
	extractedDecls := make(map[schema.Decl]types.Object)
	// for identifying duplicates
	typeUniqueness := make(map[string]tuple.Pair[types.Object, schema.Position])
	globalUniqueness := make(map[string]tuple.Pair[types.Object, schema.Position])
	for _, r := range fResults {
		fr, ok := r.(finalize.Result)
		if !ok {
			return Result{}, fmt.Errorf("unexpected schema extraction result type: %T", r)
		}
		if combined.Module == nil {
			combined.Module = &schema.Module{Name: fr.ModuleName, Comments: fr.ModuleComments}
		} else {
			if combined.Module.Name != fr.ModuleName {
				return Result{}, fmt.Errorf("unexpected schema extraction result module name: %s", fr.ModuleName)
			}
			if len(combined.Module.Comments) == 0 {
				combined.Module.Comments = fr.ModuleComments
			}
		}
		copyFailedRefs(refResults, fr.Failed)
		for decl, obj := range fr.Extracted {
			// check for duplicates and add the Decl to the module schema
			typename := common.GetDeclTypeName(decl)
			typeKey := fmt.Sprintf("%s-%s", typename, decl.GetName())
			if value, ok := typeUniqueness[typeKey]; ok && value.A != obj {
				// decls redeclared in subpackage
				combined.Errors = append(combined.Errors, schema.Errorf(decl.Position(), decl.Position().Column,
					"duplicate %s declaration for %q; already declared at %q", typename,
					combined.Module.Name+"."+decl.GetName(), value.B))
				continue
			}
			if value, ok := globalUniqueness[decl.GetName()]; ok && value.A != obj {
				combined.Errors = append(combined.Errors, schema.Errorf(decl.Position(), decl.Position().Column,
					"schema declaration with name %q already exists for module %q; previously declared at %q",
					decl.GetName(), combined.Module.Name, value.B))
			}
			typeUniqueness[typeKey] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
			globalUniqueness[decl.GetName()] = tuple.Pair[types.Object, schema.Position]{A: obj, B: decl.Position()}
			extractedDecls[decl] = obj
		}
		maps.Copy(combined.NativeNames, fr.NativeNames)
	}

	combined.Module.AddDecls(maps.Keys(extractedDecls))
	for decl, obj := range extractedDecls {
		if ta, ok := decl.(*schema.TypeAlias); ok && len(ta.Metadata) > 0 {
			fqName, err := goQualifiedNameForWidenedType(obj, ta.Metadata)
			if err != nil {
				combined.Errors = append(combined.Errors, &schema.Error{Pos: ta.Position(), EndColumn: ta.Pos.Column,
					Msg: err.Error(), Level: schema.ERROR})
				continue
			}
			refResults[schema.RefKey{Module: combined.Module.Name, Name: ta.Name}] =
				refResult{typ: widened, obj: obj, fqName: optional.Some(fqName)}
		}
		combined.NativeNames[decl] = common.GetNativeName(obj)
	}
	combined.Errors = append(combined.Errors, propagateTypeErrors(combined.Module, refResults)...)
	schema.SortErrorsByPosition(combined.Errors)
	updateVisibility(combined.Module)
	// TODO: validate schema once we have the full schema here
	return combined, nil
}

func copyFailedRefs(parsedRefs map[schema.RefKey]refResult, failedRefs map[schema.RefKey]types.Object) {
	for ref, obj := range failedRefs {
		parsedRefs[ref] = refResult{typ: failed, obj: obj}
	}
}

// updateVisibility traverses the module schema via refs and updates visibility as needed.
func updateVisibility(module *schema.Module) {
	for _, d := range module.Decls {
		if d.IsExported() {
			updateTransitiveVisibility(d, module)
		}
	}
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

// propagateTypeErrors propagates type errors to referencing nodes. This improves error messaging for the LSP client by
// surfacing errors all the way up the schema chain.
func propagateTypeErrors(
	module *schema.Module,
	refResults map[schema.RefKey]refResult,
) []*schema.Error {
	var errs []*schema.Error
	_ = schema.VisitWithParent(module, nil, func(n schema.Node, p schema.Node, next func() error) error { //nolint:errcheck
		if p == nil {
			return next()
		}
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}

		result, ok := refResults[ref.ToRefKey()]
		if !ok {
			return next()
		}

		switch result.typ {
		case failed:
			refNativeName := common.GetNativeName(result.obj)
			switch pt := p.(type) {
			case *schema.Verb:
				if pt.Request == n {
					errs = append(errs, schema.Errorf(pt.Request.Position(), pt.Request.Position().Column,
						"unsupported request type %q", refNativeName))
				}
				if pt.Response == n {
					errs = append(errs, schema.Errorf(pt.Response.Position(), pt.Response.Position().Column,
						"unsupported response type %q", refNativeName))
				}
			case *schema.Field:
				errs = append(errs, schema.Errorf(pt.Position(), pt.Position().Column, "unsupported type %q for "+
					"field %q", refNativeName, pt.Name))
			default:
				errs = append(errs, schema.Errorf(p.Position(), p.Position().Column, "unsupported type %q",
					refNativeName))
			}
		case widened:
			errs = append(errs, schema.Warnf(n.Position(), n.Position().Column, "external type %q will be "+
				"widened to Any", result.fqName.MustGet()))
		}

		return next()
	})
	return errs
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

func dependenciesBeforeIndex(idx int) []*analysis.Analyzer {
	var deps []*analysis.Analyzer
	for i := range idx {
		deps = append(deps, Extractors[i]...)
	}
	return deps
}

func simplePosToSchemaPos(pos analysis.SimplePosition) schema.Position {
	return schema.Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
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
	if nativeName == "" {
		return "", fmt.Errorf("missing Go native name in typemapped alias for %q",
			common.GetNativeName(obj))
	}
	return nativeName, nil
}
