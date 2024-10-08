package call

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

const (
	ftlPkgPath             = "github.com/TBD54566975/ftl/go-runtime/ftl"
	ftlTopicHandleTypeName = "TopicHandle"
)

// Extractor extracts all function calls.
var Extractor = common.NewExtractor("validate", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}
	var currentFunc *ast.FuncDecl
	in.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.FuncDecl:
			currentFunc = node
		case *ast.CallExpr:
			validateCallExpr(pass, node)
			if currentFunc == nil {
				return
			}
			parentFuncObj, ok := common.GetObjectForNode(pass.TypesInfo, currentFunc).Get()
			if !ok {
				return
			}
			_, fn := common.Deref[*types.Func](pass, node.Fun)
			if fn == nil {
				return
			}
			common.MarkFunctionCall(pass, parentFuncObj, fn, common.GoPosToSchemaPos(pass.Fset, node.Pos()))
		}
	})
	return common.NewExtractorResult(pass), nil
}

// validateCallExpr validates all function calls
// checks if the function call is:
// - a direct verb call to an external module
// - a direct publish call on an external module's topic
func validateCallExpr(pass *analysis.Pass, node *ast.CallExpr) {
	selExpr, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	var lhsIdent *ast.Ident
	if expr, ok := selExpr.X.(*ast.SelectorExpr); ok {
		lhsIdent = expr.Sel
	} else if ident, ok := selExpr.X.(*ast.Ident); ok {
		lhsIdent = ident
	} else {
		return
	}
	lhsObject := pass.TypesInfo.ObjectOf(lhsIdent)
	if lhsObject == nil {
		return
	}
	var lhsPkgPath string
	if pkgName, ok := lhsObject.(*types.PkgName); ok {
		lhsPkgPath = pkgName.Imported().Path()
	} else {
		lhsPkgPath = lhsObject.Pkg().Path()
	}

	var lhsIsExternal bool
	if !common.IsPathInModule(pass.Pkg, lhsPkgPath) {
		lhsIsExternal = true
	}

	if lhsType, ok := pass.TypesInfo.TypeOf(selExpr.X).(*types.Named); ok && lhsType.Obj().Pkg() != nil &&
		lhsType.Obj().Pkg().Path() == ftlPkgPath {
		// Calling a function on an FTL type
		if lhsType.Obj().Name() == ftlTopicHandleTypeName && selExpr.Sel.Name == "Publish" {
			if lhsIsExternal {
				common.Errorf(pass, node, "can not publish directly to topics in other modules")
			}
		}
		return
	}
}
