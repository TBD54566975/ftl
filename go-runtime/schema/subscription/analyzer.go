package subscription

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

// Extractor extracts subscriptions.
var Extractor = common.NewResourceDeclExtractor[*schema.Subscription]("subscription", Extract, matchFunc)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.TypeSpec) optional.Option[*schema.Subscription] {
	idxListExpr, ok := node.Type.(*ast.IndexListExpr)
	if !ok {
		common.Errorf(pass, node, "unsupported subscription type")
		return optional.None[*schema.Subscription]()
	}

	if len(idxListExpr.Indices) != 3 {
		common.Errorf(pass, node, "subscription type must have exactly three type parameters")
		return optional.None[*schema.Subscription]()
	}

	topicRef, ok := common.ExtractSimpleRefWithCasing(pass, idxListExpr.Indices[0], strcase.ToLowerCamel).Get()
	if !ok {
		common.Errorf(pass, node, "unsupported topic type; please declare topics on a separate line: `MyTopic = ftl.TopicHandle[MyEvent]`, then use `MyTopic` in the subscription")
		return optional.None[*schema.Subscription]()
	}

	subscription := &schema.Subscription{
		Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:  strcase.ToLowerCamel(node.Name.Name),
		Topic: topicRef,
	}
	common.ApplyMetadata[*schema.Subscription](pass, obj, func(md *common.ExtractedMetadata) {
		subscription.Comments = md.Comments
	})
	sinkName := getSubscribingVerbName(pass, idxListExpr.Indices[1])
	if sinkName == "" || !schema.ValidateName(sinkName) {
		common.Errorf(pass, node, "unsupported sink for subscription; please provide the generated "+
			"client corresponding to the verb sink")
		return optional.None[*schema.Subscription]()
	}
	common.MarkSubscriptionSink(pass, obj, sinkName, subscription)
	return optional.Some(subscription)
}

func getSubscribingVerbName(pass *analysis.Pass, subscriber ast.Expr) string {
	obj, ok := common.GetObjectForNode(pass.TypesInfo, subscriber).Get()
	if !ok {
		return ""
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return ""
	}
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return ""
	}
	if _, ok := named.Underlying().(*types.Signature); !ok {
		return ""
	}
	module, err := common.FtlModuleFromGoPackage(obj.Pkg().Path())
	if err != nil {
		return ""
	}
	passModule, err := common.FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return ""
	}
	if module != passModule {
		common.Errorf(pass, subscriber, "sink must be in the same module as the subscription")
		return ""
	}
	ident, ok := subscriber.(*ast.Ident)
	if !ok {
		return ""
	}
	return strings.TrimSuffix(strcase.ToLowerCamel(ident.Name), "Client")
}

func matchFunc(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	return common.GetVerbResourceType(pass, obj) == common.VerbResourceTypeSubscriptionHandle
}
