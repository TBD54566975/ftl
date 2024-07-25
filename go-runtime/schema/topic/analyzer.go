package topic

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

const (
	ftlTopicFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Topic"
)

// Extractor extracts topics.
var Extractor = common.NewCallDeclExtractor[*schema.Topic]("topic", Extract, ftlTopicFuncPath)

// expects: var NameLiteral = ftl.Topic[EventType]("name_literal")
func Extract(pass *analysis.Pass, obj types.Object, node *ast.GenDecl, callExpr *ast.CallExpr,
	callPath string) optional.Option[*schema.Topic] {
	indexExpr, ok := callExpr.Fun.(*ast.IndexExpr)
	if !ok {
		common.Errorf(pass, node, "must have an event type as a type parameter")
		return optional.None[*schema.Topic]()
	}
	typeParamType, ok := common.ExtractType(pass, indexExpr.Index).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported event type")
		return optional.None[*schema.Topic]()
	}

	topicName := common.ExtractStringLiteralArg(pass, callExpr, 0)
	expTopicName := strcase.ToLowerSnake(topicName)
	if topicName != expTopicName {
		common.Errorf(pass, node, "unsupported topic name %q, did you mean to use %q?", topicName, expTopicName)
		return optional.None[*schema.Topic]()
	}

	if len(node.Specs) > 0 {
		if t, ok := node.Specs[0].(*ast.ValueSpec); ok {
			varName := t.Names[0].Name
			expVarName := strcase.ToUpperStrippedCamel(topicName)
			if varName != expVarName {
				common.Errorf(pass, node, "unexpected topic variable name %q, did you mean %q?", varName, expVarName)
				return optional.None[*schema.Topic]()
			}
		}
	}

	topic := &schema.Topic{
		Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:  topicName,
		Event: typeParamType,
	}
	common.ApplyMetadata[*schema.Subscription](pass, obj, func(md *common.ExtractedMetadata) {
		topic.Comments = md.Comments
		topic.Export = md.IsExported
	})
	return optional.Some(topic)
}
