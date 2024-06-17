package finalize

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/maps"
)

// Analyzer aggregates the results of all extractors.
var Analyzer = &analysis.Analyzer{
	Name:             "finalizer",
	Doc:              "finalizes module schema and writes to the output destination",
	Run:              Run,
	ResultType:       reflect.TypeFor[Result](),
	RunDespiteErrors: true,
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

func Run(pass *analysis.Pass) (interface{}, error) {
	moduleName, err := common.FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return nil, err
	}
	module := &schema.Module{
		Name:     moduleName,
		Comments: extractModuleComments(pass),
	}
	result := combineExtractorResults(pass, moduleName)
	module.AddDecls(result.decls)
	return Result{
		Module:      module,
		NativeNames: result.nativeNames,
	}, nil
}

func extractModuleComments(pass *analysis.Pass) []string {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.File)(nil),
	}
	var comments []string
	in.Preorder(nodeFilter, func(n ast.Node) {
		if len(strings.Split(pass.Pkg.Path(), "/")) > 2 {
			// skip subpackages
			return
		}
		comments = common.ExtractComments(n.(*ast.File).Doc) //nolint:forcetypeassert
	})
	return comments
}

type combinedResult struct {
	decls       []schema.Decl
	nativeNames map[schema.Node]string
}

func combineExtractorResults(pass *analysis.Pass, moduleName string) combinedResult {
	nn := make(map[schema.Node]string)
	extracted := make(map[types.Object]schema.Decl)
	failed := sets.NewSet[schema.RefKey]()
	for obj, fact := range common.MergeAllFacts(pass) {
		switch f := fact.Get().(type) {
		case *common.ExtractedDecl:
			if f.Decl != nil {
				extracted[obj] = f.Decl
			}
			nn[f.Decl] = obj.Pkg().Path() + "." + obj.Name()
		case *common.FailedExtraction:
			failed.Add(schema.RefKey{Module: moduleName, Name: obj.Name()})
		}
	}
	propagateTypeErrors(pass, extracted, failed)
	return combinedResult{
		nativeNames: nn,
		decls:       maps.Values(extracted),
	}
}

// propagateTypeErrors propagates type errors to referencing nodes. This improves error messaging for the LSP client by
// surfacing errors all the way up the schema chain.
func propagateTypeErrors(pass *analysis.Pass, extracted map[types.Object]schema.Decl, failed sets.Set[schema.RefKey]) {
	for obj, sch := range extracted {
		switch t := sch.(type) {
		case *schema.Verb:
			fnt := obj.(*types.Func)             //nolint:forcetypeassert
			sig := fnt.Type().(*types.Signature) //nolint:forcetypeassert
			params := sig.Params()
			results := sig.Results()
			if hasFailedRef(t.Request, failed) {
				common.TokenErrorf(pass, params.At(1).Pos(), params.At(1).Name(),
					"unsupported request type %q", params.At(1).Type())
			}
			if hasFailedRef(t.Response, failed) {
				common.TokenErrorf(pass, results.At(0).Pos(), results.At(0).Name(),
					"unsupported response type %q", results.At(0).Type())
			}
		default:
		}
	}
}

func hasFailedRef(node schema.Node, failedRefs sets.Set[schema.RefKey]) bool {
	failed := false
	_ = schema.Visit(node, func(n schema.Node, next func() error) error {
		ref, ok := n.(*schema.Ref)
		if !ok {
			return next()
		}
		if failedRefs.Contains(ref.ToRefKey()) {
			failed = true
		}
		return next()
	})
	return failed
}
