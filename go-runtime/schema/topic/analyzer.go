package topic

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
	"github.com/alecthomas/types/optional"
)

const (
	ftlTopicFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Topic"
)

// Extractor extracts topics.
var Extractor = common.NewExtractor("topic", (*Fact)(nil), Extract)

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
		if !common.FuncPathEquals(pass, callExpr, ftlTopicFuncPath) {
			return
		}
		obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
		if !ok {
			return
		}
		if topic, ok := extractTopic(pass, node, callExpr, obj).Get(); ok {
			common.MarkSchemaDecl(pass, obj, topic)
		}
	})
	return common.NewExtractorResult(pass), nil
}

// expects: var NameLiteral = ftl.Topic[EventType]("name_literal")
func extractTopic(pass *analysis.Pass, node *ast.GenDecl, callExpr *ast.CallExpr, obj types.Object) optional.Option[*schema.Topic] {
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
