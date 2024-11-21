package topic

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
var Extractor = common.NewResourceDeclExtractor[*schema.Topic]("topic", Extract, matchFunc)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.TypeSpec) optional.Option[*schema.Topic] {
	idxExpr, ok := node.Type.(*ast.IndexExpr)
	if !ok {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	typ, ok := common.ExtractType(pass, idxExpr.Index).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	name := strcase.ToLowerCamel(node.Name.Name)
	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "topic names must be valid identifiers")
		return optional.None[*schema.Topic]()
	}

	topic := &schema.Topic{
		Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:  name,
		Event: typ,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		topic.Comments = md.Comments
		topic.Export = md.IsExported
	}
	return optional.Some(topic)
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return common.GetVerbResourceType(pass, obj) == common.VerbResourceTypeTopicHandle
}
