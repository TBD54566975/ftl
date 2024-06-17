package typealias

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"
)

// Extractor extracts type aliases to the module schema.
var Extractor = common.NewDeclExtractor[*schema.TypeAlias, *ast.TypeSpec]("typealias", (*Fact)(nil), Extract)

type Fact struct {
	value common.SchemaFactValue
}

func (t *Fact) AFact()                       {}
func (t *Fact) Set(v common.SchemaFactValue) { t.value = v }
func (t *Fact) Get() common.SchemaFactValue  { return t.value }

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
