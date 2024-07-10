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
	schType, ok := common.ExtractTypeForNode(pass, obj, node.Type, nil).Get()
	if !ok {
		return optional.None[*schema.TypeAlias]()
	}

	alias := &schema.TypeAlias{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToUpperCamel(obj.Name()),
		Type: schType,
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		alias.Comments = md.Comments
		alias.Export = md.IsExported
		alias.Metadata = md.Metadata

		if len(md.Metadata) > 0 {
			hasGoTypeMap := false
			nativeName := qualifiedNameFromSelectorExpr(pass, node.Type)
			if nativeName == "" {
				return optional.None[*schema.TypeAlias]()
			}
			for _, m := range md.Metadata {
				if mt, ok := m.(*schema.MetadataTypeMap); ok {
					if mt.Runtime != "go" {
						continue
					}
					if nativeName != mt.NativeName {
						common.Errorf(pass, node, "declared type %s in typemap does not match native type %s",
							mt.NativeName, nativeName)
						return optional.None[*schema.TypeAlias]()
					}
					hasGoTypeMap = true
				} else {
					common.Errorf(pass, node, "unexpected directive on typealias %s", m)
				}
			}

			// if this alias contains any type mappings, implicitly add a Go type mapping if not already present
			if !hasGoTypeMap {
				alias.Metadata = append(alias.Metadata, &schema.MetadataTypeMap{
					Pos:        common.GoPosToSchemaPos(pass.Fset, obj.Pos()),
					Runtime:    "go",
					NativeName: nativeName,
				})
			}
			alias.Type = &schema.Any{}
			return optional.Some(alias)
		}
	} else if _, ok := alias.Type.(*schema.Any); ok &&
		!strings.HasPrefix(qualifiedNameFromSelectorExpr(pass, node.Type), "ftl") {
		alias.Metadata = append(alias.Metadata, &schema.MetadataTypeMap{
			Pos:        common.GoPosToSchemaPos(pass.Fset, obj.Pos()),
			Runtime:    "go",
			NativeName: qualifiedNameFromSelectorExpr(pass, node.Type),
		})
		return optional.Some(alias)
	}

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
