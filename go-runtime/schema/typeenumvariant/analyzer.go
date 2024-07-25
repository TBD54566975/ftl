package typeenumvariant

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

// Extractor extracts possible type enum variants.
var Extractor = common.NewExtractor("typeenumvariant", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.TypeSpec) //nolint:forcetypeassert
		obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
		if !ok {
			return
		}
		extractEnumVariant(pass, node, obj)
	})
	return common.NewExtractorResult(pass), nil
}

func extractEnumVariant(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) {
	typ := pass.TypesInfo.TypeOf(node.Type)
	if common.IsType[*types.Interface](typ) {
		return
	}

	variant := &schema.EnumVariant{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToUpperCamel(node.Name.Name),
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		variant.Comments = md.Comments
	}
	for o := range common.GetAllFactsOfType[*common.MaybeTypeEnum](pass) {
		named, ok := pass.TypesInfo.TypeOf(node.Name).(*types.Named)
		if !ok {
			continue
		}
		iType := o.Type().Underlying().(*types.Interface) //nolint:forcetypeassert
		if !types.Implements(named, iType) {
			continue
		}

		// valueFunc is only executed if this potential variant actually makes it to the schema.
		// Executing may result in transitive schema extraction, so we only execute if necessary.
		valueFunc := func(p *analysis.Pass) optional.Option[*schema.TypeValue] {
			value, ok := common.ExtractType(p, node).Get()
			if !ok {
				return optional.None[*schema.TypeValue]()
			}
			return optional.Some(&schema.TypeValue{
				Pos:   common.GoPosToSchemaPos(p.Fset, node.Pos()),
				Value: value,
			})
		}
		common.MarkMaybeTypeEnumVariant(pass, obj, variant, o, valueFunc)
	}
}
