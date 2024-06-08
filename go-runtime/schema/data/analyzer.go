package data

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"unicode"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"
)

var (
	// Extractor extracts schema.Data to the module schema.
	Extractor = common.NewDeclExtractor[*schema.Data, *ast.TypeSpec]("data", (*Fact)(nil), Extract)

	aliasFieldTag = "json"
)

type Fact struct {
	value common.SchemaFactValue
}

func (t *Fact) AFact()                       {}
func (t *Fact) Set(v common.SchemaFactValue) { t.value = v }
func (t *Fact) Get() common.SchemaFactValue  { return t.value }

func Extract(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) optional.Option[*schema.Data] {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return optional.None[*schema.Data]()
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return optional.None[*schema.Data]()
	}
	decl, ok := extractData(pass, node.Pos(), named).Get()
	if !ok {
		return optional.None[*schema.Data]()
	}
	return optional.Some(decl)
}

func extractData(pass *analysis.Pass, pos token.Pos, named *types.Named) optional.Option[*schema.Data] {
	fset := pass.Fset
	nodePath := named.Obj().Pkg().Path()
	if !common.IsPathInPkg(pass.Pkg, nodePath) {
		return optional.None[*schema.Data]()
	}

	out := &schema.Data{
		Pos:  common.GoPosToSchemaPos(fset, pos),
		Name: strcase.ToUpperCamel(named.Obj().Name()),
	}
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, named.Obj()).Get(); ok {
		if _, ok = md.Type.(*schema.Data); !ok && md.Type != nil {
			return optional.None[*schema.Data]()
		}
		out.Comments = md.Comments
		out.Export = md.IsExported
	}

	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  common.GoPosToSchemaPos(fset, pos),
			Name: param.Obj().Name(),
		})
	}

	// If the struct is generic, we need to use the origin type to get the
	// fields.
	if named.TypeParams().Len() > 0 {
		named = named.Origin()
	}

	s, ok := named.Underlying().(*types.Struct)
	if !ok {
		return optional.None[*schema.Data]()
	}

	fieldErrors := false
	for i := range s.NumFields() {
		f := s.Field(i)
		if ft, ok := common.ExtractType(pass, f.Pos(), f.Type()).Get(); ok {
			// Check if field is exported
			if len(f.Name()) > 0 && unicode.IsLower(rune(f.Name()[0])) {
				common.TokenErrorf(pass, f.Pos(), f.Name(),
					"struct field %s must be exported by starting with an uppercase letter", f.Name())
				fieldErrors = true
			}

			// Extract the JSON tag and split it to get just the field name
			tagContent := reflect.StructTag(s.Tag(i)).Get(aliasFieldTag)
			tagParts := strings.Split(tagContent, ",")
			jsonFieldName := ""
			if len(tagParts) > 0 {
				jsonFieldName = tagParts[0]
			}

			var metadata []schema.Metadata
			if jsonFieldName != "" {
				metadata = append(metadata, &schema.MetadataAlias{
					Pos:   common.GoPosToSchemaPos(pass.Fset, pos),
					Kind:  schema.AliasKindJSON,
					Alias: jsonFieldName,
				})
			}
			out.Fields = append(out.Fields, &schema.Field{
				Pos:      common.GoPosToSchemaPos(pass.Fset, pos),
				Name:     strcase.ToLowerCamel(f.Name()),
				Type:     ft,
				Metadata: metadata,
			})
		} else {
			common.TokenErrorf(pass, f.Pos(), f.Name(), "unsupported type %q for field %q", f.Type(), f.Name())
			fieldErrors = true
		}
	}
	if fieldErrors {
		return optional.None[*schema.Data]()
	}
	return optional.Some(out)
}
