package schema

import (
	"fmt"

	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/go-runtime/schema/data"
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
		metadata.Extractor,
	},
	{
		typealias.Extractor,
		verb.Extractor,
		data.Extractor,
	},
	{
		transitive.Extractor,
	},
	{
		finalize.Analyzer,
	},
}

// Extract statically parses Go FTL module source into a schema.Module
func Extract(moduleDir string) (finalize.Result, error) {
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
		return finalize.Result{}, err
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

// the run will produce finalizer results for all packages it executes on, so we need to aggregate the results into a
// single schema
func combineAllPackageResults(results map[*analysis.Analyzer][]any, diagnostics []analysis.SimpleDiagnostic) (finalize.Result, error) {
	fResults, ok := results[finalize.Analyzer]
	if !ok {
		return finalize.Result{}, fmt.Errorf("schema extraction finalizer result not found")
	}
	combined := finalize.Result{
		NativeNames: make(map[schema.Node]string),
		Errors:      diagnosticsToSchemaErrors(diagnostics),
	}
	for _, r := range fResults {
		fr, ok := r.(finalize.Result)
		if !ok {
			return finalize.Result{}, fmt.Errorf("unexpected schema extraction result type: %T", r)
		}

		if combined.Module == nil {
			combined.Module = fr.Module
		} else {
			if combined.Module.Name != fr.Module.Name {
				return finalize.Result{}, fmt.Errorf("unexpected schema extraction result module name: %s", fr.Module.Name)
			}
			combined.Module.AddDecls(fr.Module.Decls)
		}
		maps.Copy(combined.NativeNames, fr.NativeNames)
	}
	schema.SortErrorsByPosition(combined.Errors)
	updateVisibility(combined.Module)
	// TODO: validate schema once we have the full schema here
	return combined, nil
}

func dependenciesBeforeIndex(idx int) []*analysis.Analyzer {
	var deps []*analysis.Analyzer
	for i := range idx {
		deps = append(deps, Extractors[i]...)
	}
	return deps
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

func simplePosToSchemaPos(pos analysis.SimplePosition) schema.Position {
	return schema.Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}
