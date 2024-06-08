package initialize

import (
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/packages"
)

// Analyzer prepares data prior to the schema extractor runs, e.g. loads FTL types for reference by other
// analyzers.
var Analyzer = &analysis.Analyzer{
	Name:             "initialize",
	Doc:              "loads data to be used by other analyzers in the schema extractor pass",
	Run:              Run,
	ResultType:       reflect.TypeFor[Result](),
	RunDespiteErrors: true,
}

type Result struct {
	types map[string]*types.Interface
}

// IsFtlErrorType will return true if the provided type is assertable to the `builtin.error` type.
func (r Result) IsFtlErrorType(typ types.Type) bool {
	return r.assertableToType(typ, "builtin", "error")
}

// IsContextType will return true if the provided type is assertable to the `context.Context` type.
func (r Result) IsContextType(typ types.Type) bool {
	return r.assertableToType(typ, "context", "Context")
}

func (r Result) assertableToType(typ types.Type, pkg string, name string) bool {
	ityp, ok := r.types[pkg+"."+name]
	if !ok {
		return false
	}
	return types.AssertableTo(ityp, typ)
}

func Run(pass *analysis.Pass) (interface{}, error) {
	ctxType, err := loadRef("context", "Context")
	if err != nil {
		return nil, err
	}
	errType, err := loadRef("builtin", "error")
	if err != nil {
		return nil, err
	}

	return Result{types: map[string]*types.Interface{
		"context.Context": ctxType,
		"builtin.error":   errType,
	}}, nil
}

// Lazy load the compile-time reference from a package.
func loadRef(pkg, name string) (*types.Interface, error) {
	pkgs, err := packages.Load(&packages.Config{Fset: token.NewFileSet(), Mode: packages.NeedTypes}, pkg)
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected one package, got %s",
			strings.Join(slices.Map(pkgs, func(p *packages.Package) string { return p.Name }), ", "))
	}
	obj := pkgs[0].Types.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("interface %q not found", name)
	}
	ifaceType, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("expected an interface, got %s", obj.Type())
	}
	return ifaceType, nil
}
