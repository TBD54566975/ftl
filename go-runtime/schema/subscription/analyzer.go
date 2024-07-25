package subscription

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"
)

const (
	ftlSubscriptionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Subscription"
)

// Extractor extracts subscriptions.
var Extractor = common.NewCallDeclExtractor[*schema.Subscription]("subscription", Extract, ftlSubscriptionFuncPath)

// expects: var _ = ftl.Subscription(topicHandle, "name_literal")
func Extract(pass *analysis.Pass, obj types.Object, node *ast.GenDecl, callExpr *ast.CallExpr,
	callPath string) optional.Option[*schema.Subscription] {
	var topicRef *schema.Ref
	if len(callExpr.Args) != 2 {
		common.Errorf(pass, callExpr, "subscription registration must have exactly two arguments")
		return optional.None[*schema.Subscription]()
	}
	if topicIdent, ok := callExpr.Args[0].(*ast.Ident); ok {
		// Topic is within module
		// we will find the subscription name from the string literal parameter
		object := pass.TypesInfo.ObjectOf(topicIdent)
		fact, ok := common.GetFactForObject[*common.ExtractedDecl](pass, object).Get()
		if !ok || fact.Decl == nil {
			common.Errorf(pass, callExpr, "could not find topic declaration for topic variable")
			return optional.None[*schema.Subscription]()
		}
		topic, ok := fact.Decl.(*schema.Topic)
		if !ok {
			common.Errorf(pass, callExpr, "could not find topic declaration for topic variable")
			return optional.None[*schema.Subscription]()
		}

		moduleName, err := common.FtlModuleFromGoPackage(pass.Pkg.Path())
		if err != nil {
			return optional.None[*schema.Subscription]()
		}
		topicRef = &schema.Ref{
			Module: moduleName,
			Name:   topic.Name,
		}
	} else if topicSelExp, ok := callExpr.Args[0].(*ast.SelectorExpr); ok {
		// External topic
		// we will derive subscription name from generated variable name
		moduleIdent, moduleOk := topicSelExp.X.(*ast.Ident)
		if !moduleOk {
			common.Errorf(pass, callExpr, "subscription registration must have a topic")
			return optional.None[*schema.Subscription]()
		}
		varName := topicSelExp.Sel.Name
		if varName == "" {
			common.Errorf(pass, callExpr, "subscription registration must have a topic")
			return optional.None[*schema.Subscription]()
		}
		name := strcase.ToLowerSnake(varName)
		topicRef = &schema.Ref{
			Module: moduleIdent.Name,
			Name:   name,
		}
	} else {
		common.Errorf(pass, callExpr, "subscription registration must have a topic")
		return optional.None[*schema.Subscription]()
	}

	subscription := &schema.Subscription{
		Pos:   common.GoPosToSchemaPos(pass.Fset, callExpr.Pos()),
		Name:  common.ExtractStringLiteralArg(pass, callExpr, 1),
		Topic: topicRef,
	}
	common.ApplyMetadata[*schema.Subscription](pass, obj, func(md *common.ExtractedMetadata) {
		subscription.Comments = md.Comments
	})
	return optional.Some(subscription)
}
