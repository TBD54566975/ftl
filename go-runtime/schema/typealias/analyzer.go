package typealias

import (
	"go/ast"
	"go/types"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
)

// Extractor extracts type aliases to the module schema.
var Extractor = common.NewDeclExtractor[*schema.TypeAlias, *ast.TypeSpec]("typealias", Extract)

func Extract(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) optional.Option[*schema.TypeAlias] {
	schType, ok := common.ExtractTypeForNode(pass, obj, node, nil).Get()
	if !ok {
		return optional.None[*schema.TypeAlias]()
	}
	if common.IsSelfReference(pass, obj, schType) {
		return optional.None[*schema.TypeAlias]()
	}
	alias := &schema.TypeAlias{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToUpperCamel(obj.Name()),
		Type: schType,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		if _, ok := md.Type.(*schema.TypeAlias); !ok {
			return optional.None[*schema.TypeAlias]()
		}
		alias.Comments = md.Comments
		alias.Export = md.IsExported
	}
	return optional.Some(alias)
}
