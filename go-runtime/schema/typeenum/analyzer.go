package typeenum

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

// Extractor extracts possible type enum discriminators.
//
// All named interfaces are marked as possible type enum discriminators and subsequent extractors determine if they are
// part of an enum.
var Extractor = common.NewExtractor("typeenum", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.TypeSpec) //nolint:forcetypeassert

		iType, ok := pass.TypesInfo.TypeOf(node.Type).Underlying().(*types.Interface)
		if !ok {
			return
		}

		obj, ok := common.GetObjectForNode(pass.TypesInfo, node).Get()
		if !ok {
			return
		}

		enum := &schema.Enum{
			Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
			Name: strcase.ToUpperCamel(node.Name.Name),
		}
		common.ApplyMetadata[*schema.Enum](pass, obj, func(md *common.ExtractedMetadata) {
			enum.Comments = md.Comments
			enum.Export = md.IsExported

			if _, ok := md.Type.(*schema.Enum); ok {
				if iType.NumMethods() == 0 {
					common.Errorf(pass, node, "enum discriminator %q must define at least one method", node.Name.Name)
					return
				}
				for i := range iType.NumMethods() {
					m := iType.Method(i)
					if m.Exported() {
						common.Errorf(pass, node, "enum discriminator %q cannot contain exported methods",
							node.Name.Name)
						return
					}
				}
				common.MarkNeedsExtraction(pass, obj)
			}
		})
		if iType.NumMethods() > 0 {
			common.MarkMaybeTypeEnum(pass, obj, enum)
		}
	})
	return common.NewExtractorResult(pass), nil
}
