package configsecret

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/internal/schema"
)

const (
	ftlConfigFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Config"
	ftlSecretFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Secret" //nolint:gosec
)

// Extractor extracts configs and secrets.
var Extractor = common.NewExtractor("configsecret", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.GenDecl) //nolint:forcetypeassert
		callExpr, ok := common.CallExprFromVar(node).Get()
		if !ok {
			return
		}
		obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
		if !ok {
			return
		}
		_, fn := common.Deref[*types.Func](pass, callExpr.Fun)
		if fn == nil {
			return
		}
		var comments []string
		if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
			comments = md.Comments
		}
		var decl optional.Option[schema.Decl]
		switch fn.FullName() {
		case ftlSecretFuncPath:
			decl = extractConfigSecret[*schema.Secret](pass, callExpr, comments)
		case ftlConfigFuncPath:
			decl = extractConfigSecret[*schema.Config](pass, callExpr, comments)
		}
		if d, ok := decl.Get(); ok {
			common.MarkSchemaDecl(pass, obj, d)
		}
	})
	return common.NewExtractorResult(pass), nil
}

func extractConfigSecret[T schema.Decl](
	pass *analysis.Pass,
	node *ast.CallExpr,
	comments []string,
) optional.Option[schema.Decl] {
	name := common.ExtractStringLiteralArg(pass, node, 0)
	if name == "" {
		return optional.None[schema.Decl]()
	}
	var t T
	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "%s names must be valid identifiers", common.GetDeclTypeName(t))
		return optional.None[schema.Decl]()
	}

	index := node.Fun.(*ast.IndexExpr) //nolint:forcetypeassert
	// Type parameter
	st, ok := common.ExtractType(pass, index.Index).Get()
	if !ok {
		common.Errorf(pass, index.Index, "config is unsupported type")
		return optional.None[schema.Decl]()
	}

	var decl schema.Decl
	switch any(t).(type) {
	case *schema.Config:
		decl = &schema.Config{
			Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
			Comments: comments,
			Name:     name,
			Type:     st,
		}
	case *schema.Secret:
		decl = &schema.Secret{
			Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
			Comments: comments,
			Name:     name,
			Type:     st,
		}
	default:
		return optional.None[schema.Decl]()
	}

	return optional.Some(decl)
}
