package data

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

var (
	// Extractor extracts schema.Data to the module schema.
	Extractor = common.NewDeclExtractor[*schema.Data, *ast.TypeSpec]("data", Extract)

	aliasFieldTag = "json"
)

func Extract(pass *analysis.Pass, node *ast.TypeSpec, obj types.Object) optional.Option[*schema.Data] {
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return optional.None[*schema.Data]()
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return optional.None[*schema.Data]()
	}
	decl, ok := extractData(pass, node, named).Get()
	if !ok {
		return optional.None[*schema.Data]()
	}
	return optional.Some(decl)
}

func extractData(pass *analysis.Pass, node *ast.TypeSpec, named *types.Named) optional.Option[*schema.Data] {
	out := &schema.Data{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToUpperCamel(node.Name.Name),
	}
	common.ApplyMetadata[*schema.Data](pass, named.Obj(), func(md *common.ExtractedMetadata) {
		out.Comments = md.Comments
		out.Export = md.IsExported
	})
	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
			Name: param.Obj().Name(),
		})
	}

	structType, ok := node.Type.(*ast.StructType)
	if !ok {
		return optional.None[*schema.Data]()
	}

	fieldErrors := false
	for i := range structType.Fields.List {
		f := structType.Fields.List[i]
		var name string
		if len(f.Names) > 0 && len(f.Names[0].Name) > 0 {
			name = f.Names[0].Name
		}
		if name == "" {
			common.Errorf(pass, f, "anonymous fields are not supported")
			fieldErrors = true
			continue
		}
		if ft, ok := common.ExtractType(pass, f.Type).Get(); ok {
			// Check if field is exported
			if unicode.IsLower(rune(name[0])) {
				name = f.Names[0].Name
				common.TokenErrorf(pass, f.Pos(), name, "struct field %s must be exported by starting with an "+
					"uppercase letter", name)
				fieldErrors = true
			}

			// Extract the JSON tag and split it to get just the field name
			var metadata []schema.Metadata
			if tag := f.Tag; tag != nil {
				jsonFieldName, err := parseTag(pass, f, tag, aliasFieldTag)
				if err != nil {
					fieldErrors = true
					continue
				}

				if jsonFieldName != "" {
					metadata = append(metadata, &schema.MetadataAlias{
						Pos:   common.GoPosToSchemaPos(pass.Fset, node.Pos()),
						Kind:  schema.AliasKindJSON,
						Alias: jsonFieldName,
					})
				}
			}
			out.Fields = append(out.Fields, &schema.Field{
				Pos:      common.GoPosToSchemaPos(pass.Fset, node.Pos()),
				Name:     strcase.ToLowerCamel(name),
				Type:     ft,
				Metadata: metadata,
			})
		} else {
			common.TokenErrorf(pass, f.Pos(), name, "unsupported type %q for field %q", pass.TypesInfo.TypeOf(f.Type), name)
			fieldErrors = true
		}
	}
	if fieldErrors {
		return optional.None[*schema.Data]()
	}
	return optional.Some(out)
}

func parseTag(pass *analysis.Pass, f *ast.Field, tag *ast.BasicLit, fieldTag string) (string, error) {
	unquoted, err := strconv.Unquote(tag.Value)
	if err != nil {
		common.Wrapf(pass, f, err, "failed to unquote tag value %q", tag.Value)
		return "", fmt.Errorf("failed to unquote tag value: %w", err)
	}
	tagContent := reflect.StructTag(unquoted).Get(fieldTag)
	tagParts := strings.Split(tagContent, ",")
	jsonFieldName := ""
	if len(tagParts) > 0 {
		jsonFieldName = tagParts[0]
	}
	return jsonFieldName, nil
}
