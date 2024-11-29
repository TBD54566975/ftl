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
	idxListExpr, ok := node.Type.(*ast.IndexListExpr)
	if !ok || len(idxListExpr.Indices) != 2 {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	// extract event type
	eventTypeExpr, ok := idxListExpr.Indices[0].(*ast.Ident)
	if !ok {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}
	typ, ok := common.ExtractType(pass, eventTypeExpr).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported topic type")
		return optional.None[*schema.Topic]()
	}

	// extract name
	name := strcase.ToLowerCamel(node.Name.Name)
	if !schema.ValidateName(name) {
		common.Errorf(pass, node, "topic names must be valid identifiers")
		return optional.None[*schema.Topic]()
	}

	// topic fact
	topic := &schema.Topic{
		Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:  name,
		Event: typ,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		topic.Comments = md.Comments
		topic.Export = md.IsExported
	}

	// mapper fact
	mapperObj, ok := common.GetObjectForNode(pass.TypesInfo, idxListExpr.Indices[1]).Get()
	if !ok {
		common.Errorf(pass, node, "could not find type for topic partition mapper")
		return optional.None[*schema.Topic]()
	}

	associatedExprs := []ast.Expr{}
	switch mapper := idxListExpr.Indices[1].(type) {
	case *ast.IndexExpr:
		associatedExprs = append(associatedExprs, mapper.Index)

	case *ast.IndexListExpr:
		associatedExprs = append(associatedExprs, mapper.Indices...)

	default:
	}
	associatedMapperObjs := []types.Object{}
	for _, expr := range associatedExprs {
		associatedObj, ok := common.GetObjectForNode(pass.TypesInfo, expr).Get()
		if !ok {
			common.Errorf(pass, node, "could not find associated type for topic partition mapper")
			return optional.None[*schema.Topic]()
		}
		associatedMapperObjs = append(associatedMapperObjs, associatedObj)
	}

	common.MarkTopicMapper(pass, mapperObj, associatedMapperObjs, topic)

	return optional.Some(topic)
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return common.GetVerbResourceType(pass, obj) == common.VerbResourceTypeTopicHandle
}
