package finalize

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
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
	ModuleName     string
	ModuleComments []string

	// Extracted contains all objects successfully extracted to schema.Decls.
	Extracted map[schema.Decl]types.Object
	// Failed contains all objects that failed extraction.
	Failed map[schema.RefKey]types.Object
	// Native names that can't be derived outside of the analysis pass.
	NativeNames map[schema.Node]string
	// FunctionCalls contains all function calls; key is the parent function, value is the called functions.
	FunctionCalls map[schema.Position]FunctionCall
}

type FunctionCall struct {
	Caller types.Object
	Callee types.Object
}

func Run(pass *analysis.Pass) (interface{}, error) {
	moduleName, err := common.FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return nil, err
	}
	extracted := make(map[schema.Decl]types.Object)
	failed := make(map[schema.RefKey]types.Object)
	// for identifying duplicates
	declKeys := make(map[string]types.Object)
	nativeNames := make(map[schema.Node]string)
	for obj, fact := range common.GetAllFactsExtractionStatus(pass) {
		switch f := fact.(type) {
		case *common.ExtractedDecl:
			if existing, ok := declKeys[f.Decl.String()]; ok && existing != obj && obj.Pkg().Path() == pass.Pkg.Path() {
				common.NoEndColumnErrorf(pass, obj.Pos(), "duplicate %s declaration for %q; already declared at %q",
					common.GetDeclTypeName(f.Decl), moduleName+"."+f.Decl.GetName(), common.GoPosToSchemaPos(pass.Fset, existing.Pos()))
				continue
			}
			if f.Decl != nil && pass.Pkg.Path() == obj.Pkg().Path() {
				extracted[f.Decl] = obj
				declKeys[f.Decl.String()] = obj
				nativeNames[f.Decl] = common.GetNativeName(obj)
			}
		case *common.FailedExtraction:
			failed[schema.RefKey{Module: moduleName, Name: strcase.ToUpperCamel(obj.Name())}] = obj
		}
	}
	for obj, facts := range common.GetAllFactsOfType[*common.MaybeTypeEnumVariant](pass) {
		for _, fact := range facts {
			nativeNames[fact.Variant] = common.GetNativeName(obj)
		}
	}
	for obj, facts := range common.GetAllFactsOfType[*common.IncludeNativeName](pass) {
		for _, fact := range facts {
			nativeNames[fact.Node] = common.GetNativeName(obj)
		}
	}
	return Result{
		ModuleName:     moduleName,
		ModuleComments: extractModuleComments(pass),
		Extracted:      extracted,
		Failed:         failed,
		NativeNames:    nativeNames,
		FunctionCalls:  getFunctionCalls(pass),
	}, nil
}

func getFunctionCalls(pass *analysis.Pass) map[schema.Position]FunctionCall {
	fnCalls := make(map[schema.Position]FunctionCall)
	for caller, calls := range common.GetAllFactsOfType[*common.FunctionCall](pass) {
		for _, fnCall := range calls {
			fnCalls[fnCall.Position] = FunctionCall{
				Caller: caller,
				Callee: fnCall.Callee,
			}
		}
	}
	return fnCalls
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
