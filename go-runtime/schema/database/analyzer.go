package database

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"
)

const ftlPostgresDBFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.PostgresDatabase"

// Extractor extracts databases to the module schema.
var Extractor = common.NewCallDeclExtractor[*schema.Database]("database", Extract, ftlPostgresDBFuncPath)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.GenDecl, callExpr *ast.CallExpr,
	callPath string) optional.Option[*schema.Database] {
	var comments []string
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		comments = md.Comments
	}
	if callPath == ftlPostgresDBFuncPath {
		return extractDatabase(pass, callExpr, schema.PostgresDatabaseType, comments)
	}
	return optional.None[*schema.Database]()
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
