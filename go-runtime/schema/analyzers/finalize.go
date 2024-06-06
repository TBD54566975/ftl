package analyzers

import (
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"golang.org/x/exp/maps"
)

// Extractors is a list of all schema extractors that must run.
var Extractors = []*analysis.Analyzer{
	TypeAliasExtractor,
}

// Finalizer aggregates the results of all extractors.
var Finalizer = &analysis.Analyzer{
	Name:             "finalizer",
	Doc:              "finalizes module schema and writes to the output destination",
	Run:              runFinalizer,
	Requires:         Extractors,
	ResultType:       reflect.TypeFor[ExtractResult](),
	RunDespiteErrors: true,
}

// ExtractResult contains the final schema extraction result.
type ExtractResult struct {
	// Module is the extracted module schema.
	Module *schema.Module
	// NativeNames maps schema nodes to their native Go names.
	NativeNames map[schema.Node]string
	// Errors is a list of errors encountered during schema extraction.
	Errors []*schema.Error
}

func runFinalizer(pass *analysis.Pass) (interface{}, error) {
	module, nativeNames, err := buildModuleSchema(pass)
	if err != nil {
		return nil, fmt.Errorf("could not process schema extraction results")
	}

	// TODO: validate schema once we have the full schema here

	return ExtractResult{
		Module:      module,
		NativeNames: nativeNames,
	}, nil
}

// buildModuleSchema aggregates the results of all extractors.
func buildModuleSchema(pass *analysis.Pass) (*schema.Module, NativeNames, error) {
	moduleName, err := ftlModuleFromGoModule(pass.Pkg.Path())
	if err != nil {
		return nil, nil, err
	}
	module := &schema.Module{Name: moduleName}
	nn := NativeNames{}
	for _, e := range Extractors {
		r, ok := pass.ResultOf[e].(result)
		if !ok {
			return nil, nil, fmt.Errorf("failed to extract result of %s", e.Name)
		}
		module.AddDecls(r.decls)
		maps.Copy(nn, r.nativeNames)
	}
	updateVisibility(module)
	return module, nn, nil
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
