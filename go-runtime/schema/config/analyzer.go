package config

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

// Extractor extracts topics.
var Extractor = common.NewResourceDeclExtractor[*schema.Config]("config", Extract, matchFunc)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.TypeSpec) optional.Option[*schema.Config] {
	idxExpr, ok := node.Type.(*ast.IndexExpr)
	if !ok {
		common.Errorf(pass, node, "unsupported config type")
		return optional.None[*schema.Config]()
	}

	typ, ok := common.ExtractType(pass, idxExpr.Index).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported config type")
		return optional.None[*schema.Config]()
	}

	name := strcase.ToLowerCamel(node.Name.Name)
	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "config names must be valid identifiers")
		return optional.None[*schema.Config]()
	}

	cfg := &schema.Config{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: name,
		Type: typ,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		cfg.Comments = md.Comments
	}
	return optional.Some(cfg)
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return common.GetVerbResourceType(pass, obj) == common.VerbResourceTypeConfig
}
