package finalize

import (
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
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
}

func Run(pass *analysis.Pass) (interface{}, error) {
	moduleName, err := common.FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return nil, err
	}
	extracted := make(map[schema.Decl]types.Object)
	failed := make(map[schema.RefKey]types.Object)
	for obj, fact := range common.MergeAllFacts(pass) {
		switch f := fact.Get().(type) {
		case *common.ExtractedDecl:
			if f.Decl != nil {
				extracted[f.Decl] = obj
			}
		case *common.FailedExtraction:
			failed[schema.RefKey{Module: moduleName, Name: strcase.ToUpperCamel(obj.Name())}] = obj
		}
	}
	return Result{
		ModuleName:     moduleName,
		ModuleComments: extractModuleComments(pass),
		Extracted:      extracted,
		Failed:         failed,
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
