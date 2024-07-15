package database

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
	"github.com/alecthomas/types/optional"
)

const ftlPostgresDBFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.PostgresDatabase"

// Extractor extracts databases to the module schema.
var Extractor = common.NewExtractor("database", (*Fact)(nil), Extract)

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

		_, fn := common.Deref[*types.Func](pass, callExpr.Fun)
		if fn == nil {
			return
		}

		obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
		if !ok {
			return
		}

		var comments []string
		if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
			comments = md.Comments
		}

		var decl optional.Option[*schema.Database]
		if fn.FullName() == ftlPostgresDBFuncPath {
			decl = extractDatabase(pass, callExpr, schema.PostgresDatabaseType, comments)
		}

		if d, ok := decl.Get(); ok {
			common.MarkSchemaDecl(pass, obj, d)
		}
	})

	return common.NewExtractorResult(pass), nil
}

func extractDatabase(
	pass *analysis.Pass,
	node *ast.CallExpr,
	dbType string,
	comments []string,
) optional.Option[*schema.Database] {
	name := common.ExtractStringLiteralArg(pass, node, 0)
	if name == "" {
		return optional.None[*schema.Database]()
	}

	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "invalid database name %q", name)
		return optional.None[*schema.Database]()
	}

	return optional.Some(&schema.Database{
		Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Comments: comments,
		Name:     name,
		Type:     dbType,
	})
}
