package call

import (
	"go/ast"
	"go/types"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl-golang-tools/go/analysis/passes/inspect"
	"github.com/block/ftl-golang-tools/go/ast/inspector"

	"github.com/block/ftl/go-runtime/schema/common"
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
	lhsPkgPath, aliased := pkgPathFromObject(lhsObject)
	// if the lhsObject isn't aliased (e.g. type MyTopic = TopicHandle[MyEvent]), then it's not one of our generated
	// types and we can't reliably evaluate which module it belongs to here
	//
	// rely on runtime validation in this case
	if lhsPkgPath == "" || !aliased {
		return
	}

	if selExpr.Sel.Name == "Publish" && isTopicHandleType(pass.TypesInfo.TypeOf(selExpr.X)) &&
		!common.IsPathInModule(pass.Pkg, lhsPkgPath) {
		common.Errorf(pass, node, "can not publish directly to topics in other modules")
	}
}

func isTopicHandleType(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Alias:
		return isTopicHandleType(t.Rhs())
	case *types.Named:
		return t.Obj().Pkg().Path()+"."+t.Obj().Name() == common.FtlTopicHandlePath
	}
	return false
}

func pkgPathFromObject(obj types.Object) (pkgPath string, wasAliased bool) {
	switch t := obj.(type) {
	case *types.Var:
		return pkgPathFromType(t.Type())
	case *types.PkgName:
		return t.Imported().Path(), false
	default:
		return obj.Pkg().Path(), false
	}
}

func pkgPathFromType(typ types.Type) (pkgPath string, wasAliased bool) {
	switch tt := typ.(type) {
	case *types.Alias:
		return tt.Obj().Pkg().Path(), true
	case *types.Named:
		return tt.Obj().Pkg().Path(), false
	default:
		return "", false
	}
}
