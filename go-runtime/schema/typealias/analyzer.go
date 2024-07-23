package typealias

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
)

// Extractor extracts type aliases to the module schema.
var Extractor = common.NewDeclExtractor[*schema.TypeAlias, *ast.TypeSpec]("typealias", Extract)

func Extract(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) optional.Option[*schema.TypeAlias] {
	alias := &schema.TypeAlias{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToUpperCamel(obj.Name()),
	}
	var hasGoTypeMapping bool
	common.ApplyMetadata[*schema.TypeAlias](pass, obj, func(md *common.ExtractedMetadata) {
		alias.Comments = md.Comments
		alias.Export = md.IsExported
		alias.Metadata = md.Metadata

		if len(md.Metadata) > 0 {
			nativeName := qualifiedNameFromSelectorExpr(pass, node.Type)
			if nativeName == "" {
				return
			}
			for _, m := range md.Metadata {
				if mt, ok := m.(*schema.MetadataTypeMap); ok {
					if mt.Runtime != "go" {
						continue
					}
					hasGoTypeMapping = true
					if nativeName != mt.NativeName {
						common.Errorf(pass, node, "declared type %s in typemap does not match native type %s",
							mt.NativeName, nativeName)
						return
					}
				} else {
					common.Errorf(pass, node, "unexpected directive on typealias %s", m)
				}
			}
		}
	})

	// if widening an external type, implicitly add a Go type mapping if one does not exist
	if nn := qualifiedNameFromSelectorExpr(pass, node.Type); nn != "" && common.IsExternalType(nn) {
		alias.Type = &schema.Any{}
		if !hasGoTypeMapping {
			alias.Metadata = append(alias.Metadata, &schema.MetadataTypeMap{
				Pos:        common.GoPosToSchemaPos(pass.Fset, obj.Pos()),
				Runtime:    "go",
				NativeName: nn,
			})
		}
		return optional.Some(alias)
	}

	schType, ok := common.ExtractType(pass, node.Type).Get()
	if !ok {
		return optional.None[*schema.TypeAlias]()
	}
	alias.Type = schType

	// type aliases must have an underlying type, and the type cannot be a reference to the alias itself.
	if common.IsSelfReference(pass, obj, schType) {
		return optional.None[*schema.TypeAlias]()
	}

	return optional.Some(alias)
}

func qualifiedNameFromSelectorExpr(pass *analysis.Pass, node ast.Node) string {
	se, ok := node.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := se.X.(*ast.Ident)
	if !ok {
		return ""
	}
	for _, im := range pass.Pkg.Imports() {
		if im.Name() != ident.Name {
			continue
		}
		fqName := im.Path()
		if parts := strings.Split(im.Path(), "/"); parts[len(parts)-1] != ident.Name {
			// if package differs from the directory name, add the package name to the fqName
			fqName = fqName + "." + ident.Name
		}
		return fqName + "." + se.Sel.Name
	}
	return ""
}
