package enum

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/block/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/common"
)

// Extractor extracts enums to the module schema.
var Extractor = common.NewDeclExtractor[*schema.Enum, *ast.TypeSpec]("enums", Extract)

func Extract(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) optional.Option[*schema.Enum] {
	valueVariants := findValueEnumVariants(pass, obj)
	if facts := common.GetFactsForObject[*common.MaybeTypeEnumVariant](pass, obj); len(facts) > 0 && len(valueVariants) > 0 {
		for _, te := range facts {
			common.TokenErrorf(pass, obj.Pos(), obj.Name(), "%q is a value enum and cannot be tagged as a variant of type enum %q directly",
				obj.Name(), te.Parent.Name())
		}
	}

	// type enum
	if discriminator, ok := common.GetFactForObject[*common.MaybeTypeEnum](pass, obj).Get(); ok {
		if len(valueVariants) > 0 {
			common.Errorf(pass, node, "type %q cannot be both a type and value enum", obj.Name())
			return optional.None[*schema.Enum]()
		}

		e := discriminator.Enum
		e.Variants = findTypeValueVariants(pass, obj)
		slices.SortFunc(e.Variants, func(a, b *schema.EnumVariant) int {
			return strings.Compare(a.Name, b.Name)
		})
		return optional.Some(e)
	}

	// value enum
	if len(valueVariants) == 0 {
		return optional.None[*schema.Enum]()
	}

	typ, ok := common.ExtractType(pass, node).Get()
	if !ok {
		return optional.None[*schema.Enum]()
	}

	e := &schema.Enum{
		Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name:     strcase.ToUpperCamel(obj.Name()),
		Variants: valueVariants,
		Type:     typ,
	}
	common.ApplyMetadata[*schema.Enum](pass, obj, func(md *common.ExtractedMetadata) {
		e.Comments = md.Comments
		e.Export = md.IsExported
	})
	return optional.Some(e)

}

func findValueEnumVariants(pass *analysis.Pass, obj types.Object) []*schema.EnumVariant {
	var variants []*schema.EnumVariant
	for o, facts := range common.GetAllFactsOfType[*common.MaybeValueEnumVariant](pass) {
		// there shouldn't be more than one of this type of fact on an object, but even if there are,
		// we don't care. We just need to know if there are any.
		if len(facts) < 1 {
			continue
		}
		fact := facts[0]

		if fact.Type == obj && validateVariant(pass, o, fact.Variant) {
			variants = append(variants, fact.Variant)
		}
	}
	slices.SortFunc(variants, func(a, b *schema.EnumVariant) int {
		return strings.Compare(a.Name, b.Name)
	})
	return variants
}

func validateVariant(pass *analysis.Pass, obj types.Object, variant *schema.EnumVariant) bool {
	for _, facts := range common.GetAllFactsOfType[*common.ExtractedDecl](pass) {
		if len(facts) < 1 {
			continue
		}
		fact := facts[0]

		if fact.Decl == nil {
			continue
		}
		existingEnum, ok := fact.Decl.(*schema.Enum)
		if !ok {
			continue
		}
		for _, existingVariant := range existingEnum.Variants {
			if existingVariant.Name == variant.Name && common.GoPosToSchemaPos(pass.Fset, obj.Pos()) != existingVariant.Pos {
				common.TokenErrorf(pass, obj.Pos(), obj.Name(), "enum variant %q conflicts with existing enum "+
					"variant of %q at %q", variant.Name, existingEnum.GetName(), existingVariant.Pos)
				return false
			}
		}
	}
	return true
}

func findTypeValueVariants(pass *analysis.Pass, obj types.Object) []*schema.EnumVariant {
	var variants []*schema.EnumVariant
	for vObj, facts := range common.GetAllFactsOfType[*common.MaybeTypeEnumVariant](pass) {
		// there shouldn't be more than one of this type of fact on an object, but even if there are,
		// we don't care. We just need to know if there are any.
		if len(facts) < 1 {
			continue
		}
		fact := facts[0]

		if fact.Parent != obj {
			continue
		}
		// extract variant type here rather than in the `typeenumvariant` extractor so that we only
		// call `common.ExtractType` if the enum/variant is actually part of the schema.
		//
		// the call to common.ExtractType sometimes results in transitive extraction, which we don't want during
		// the initial pass marking all *possible* variants, as some may never be used.
		value, ok := fact.GetValue(pass).Get()
		if !ok {
			common.NoEndColumnErrorf(pass, vObj.Pos(), "invalid type for enum variant %q", fact.Variant.Name)
			continue
		}
		fact.Variant.Value = value
		variants = append(variants, fact.Variant)

	}
	return variants
}
