package analysis

import (
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/alecthomas/types/optional"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"reflect"
	"strings"
	"unicode"
)

func GoPosToSchemaPos(fset *token.FileSet, pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func GoNodePosToSchemaPos(fset *token.FileSet, node ast.Node) (schema.Position, int) {
	p := fset.Position(node.Pos())
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}, fset.Position(node.End()).Column
}

func ExtractComments(doc *ast.CommentGroup) []string {
	comments := []string{}
	if doc := doc.Text(); doc != "" {
		comments = strings.Split(strings.TrimSpace(doc), "\n")
	}
	return comments
}

func ExtractType(pass *analysis.Pass, pos token.Pos, tnode types.Type, isExported bool) (optional.Option[schema.Type], []*schema.Error) {
	fset := pass.Fset
	var scherrs []*schema.Error
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return optional.Some[schema.Type](&schema.Ref{Pos: GoPosToSchemaPos(fset, pos), Name: tparam.Obj().Id()}), scherrs
	}

	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			typ, errs := ExtractRef(pass, pos, named)
			return typ, append(scherrs, errs...)
		}
		switch underlying.Kind() {
		case types.String:
			return optional.Some[schema.Type](&schema.String{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		case types.Int, types.Int64:
			return optional.Some[schema.Type](&schema.Int{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		case types.Bool:
			return optional.Some[schema.Type](&schema.Bool{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		case types.Float64:
			return optional.Some[schema.Type](&schema.Float{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		default:
			return optional.None[schema.Type](), scherrs
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			return optional.None[schema.Type](), append(scherrs, noEndColumnErrorf(fset, pos, "expected named type but got %s", tnode))
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return optional.Some[schema.Type](&schema.Time{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Unit":
			return optional.Some[schema.Type](&schema.Unit{Pos: GoPosToSchemaPos(fset, pos)}), scherrs

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Option":
			typ, errs := ExtractType(pass, pos, named.TypeArgs().At(0), isExported)
			scherrs = append(scherrs, errs...)
			if underlying, ok := typ.Get(); ok {
				return optional.Some[schema.Type](&schema.Optional{Pos: GoPosToSchemaPos(pass.Fset, pos), Type: underlying}), scherrs
			}
			return optional.None[schema.Type](), scherrs

		default:
			nodePath := named.Obj().Pkg().Path()
			if !isPathInPkg(pass.Pkg, nodePath) && !strings.HasPrefix(nodePath, "ftl/") {
				return optional.None[schema.Type](), append(scherrs,
					noEndColumnErrorf(fset, pos, "unsupported external type %s", nodePath+"."+named.Obj().Name()))
			}
			if ref, ok := ExtractData(pctx, pos, tnode, isExported).Get(); ok {
				return optional.Some[schema.Type](ref)
			}
			return optional.None[schema.Type](), scherrs
		}

	case *types.Map:
		m, errs := extractMap(pass, pos, underlying, isExported)
		scherrs = append(scherrs, errs...)
		return m, scherrs

	case *types.Slice:
		s, errs := extractSlice(pass, pos, underlying, isExported)
		scherrs = append(scherrs, errs...)
		return s, scherrs

	case *types.Interface:
		if underlying.String() == "any" {
			return optional.Some[schema.Type](&schema.Any{Pos: GoPosToSchemaPos(fset, pos)}), scherrs
		}
		if named, ok := tnode.(*types.Named); ok {
			typ, errs := ExtractRef(pass, pos, named)
			return typ, append(scherrs, errs...)
		}
		return optional.None[schema.Type](), scherrs

	default:
		return optional.None[schema.Type](), scherrs
	}
}

func ExtractData(pass *analysis.Pass, pos token.Pos, tnode types.Type, isExported bool) (optional.Option[*schema.Ref], []*schema.Error) {
	var scherrs []*schema.Error
	fset := pass.Fset
	named, ok := tnode.(*types.Named)
	if !ok {
		return optional.None[*schema.Ref](), append(scherrs,
			noEndColumnErrorf(fset, pos, "expected named type but got %s", tnode))
	}
	nodePath := named.Obj().Pkg().Path()
	if !isPathInPkg(pass.Pkg, nodePath) {
		destModule, ok := ftlModuleFromGoModule(nodePath).Get()
		if !ok {
			return optional.None[*schema.Ref](), append(scherrs,
				tokenErrorf(fset, pos, nodePath, "struct declared in non-FTL module %s", nodePath))
		}
		dataRef := &schema.Ref{
			Pos:    GoPosToSchemaPos(fset, pos),
			Module: destModule,
			Name:   named.Obj().Name(),
		}
		for i := range named.TypeArgs().Len() {
			maybeTypeArg, errs := ExtractType(pass, pos, named.TypeArgs().At(i), isExported)
			scherrs = append(scherrs, errs...)
			typeArg, ok := maybeTypeArg.Get()
			if !ok {
				scherrs = append(scherrs, tokenErrorf(fset, pos, named.TypeArgs().At(i).String(),
					"unsupported type %q for type argument", named.TypeArgs().At(i)))
				continue
			}

			// Fully qualify the Ref if needed
			if ref, okArg := typeArg.(*schema.Ref); okArg {
				if ref.Module == "" {
					ref.Module = destModule
				}
				typeArg = ref
			}
			dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
		}
		return optional.Some[*schema.Ref](dataRef), scherrs
	}

	out := &schema.Data{
		Pos:    GoPosToSchemaPos(fset, pos),
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
		Export: isExported,
	}
	pctx.nativeNames[out] = named.Obj().Name()
	dataRef := &schema.Ref{
		Pos:    GoPosToSchemaPos(fset, pos),
		Module: pctx.module.Name,
		Name:   out.Name,
	}
	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  GoPosToSchemaPos(fset, pos),
			Name: param.Obj().Name(),
		})
		typeArgs := named.TypeArgs()
		if typeArgs == nil {
			continue
		}
		maybeTypeArg, errs := ExtractType(pass, pos, typeArgs.At(i), isExported)
		scherrs = append(scherrs, errs...)
		typeArg, ok := maybeTypeArg.Get()
		if !ok {
			scherrs = append(scherrs, tokenErrorf(fset, pos, typeArgs.At(i).String(),
				"unsupported type %q for type argument", typeArgs.At(i)))
			continue
		}
		dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
	}

	// If the struct is generic, we need to use the origin type to get the
	// fields.
	if named.TypeParams().Len() > 0 {
		named = named.Origin()
	}

	// Find type declaration so we can extract comments.
	namedPos := named.Obj().Pos()
	pkg, path, _ := pctx.pathEnclosingInterval(namedPos, namedPos)
	if pkg != nil {
		for i := len(path) - 1; i >= 0; i-- {
			// We have to check both the type spec and the gen decl because the
			// type could be declared as either "type Foo struct { ... }" or
			// "type ( Foo struct { ... } )"
			switch path := path[i].(type) {
			case *ast.TypeSpec:
				if path.Doc != nil {
					out.Comments = ExtractComments(path.Doc)
				}
			case *ast.GenDecl:
				if path.Doc != nil {
					out.Comments = ExtractComments(path.Doc)
				}
			}
		}
	}

	s, ok := named.Underlying().(*types.Struct)
	if !ok {
		return optional.None[*schema.Ref](), append(scherrs,
			tokenErrorf(fset, pos, named.String(), "expected struct but got %s", named))
	}

	fieldErrors := false
	for i := range s.NumFields() {
		f := s.Field(i)
		if ft, ok := visitType(pctx, f.Pos(), f.Type(), isExported).Get(); ok {
			// Check if field is exported
			if len(f.Name()) > 0 && unicode.IsLower(rune(f.Name()[0])) {
				pctx.errors.add(tokenErrorf(f.Pos(), f.Name(),
					"struct field %s must be exported by starting with an uppercase letter", f.Name()))
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
					Pos:   goPosToSchemaPos(f.Pos()),
					Kind:  schema.AliasKindJSON,
					Alias: jsonFieldName,
				})
			}
			out.Fields = append(out.Fields, &schema.Field{
				Pos:      goPosToSchemaPos(f.Pos()),
				Name:     strcase.ToLowerCamel(f.Name()),
				Type:     ft,
				Metadata: metadata,
			})
		} else {
			pctx.errors.add(tokenErrorf(f.Pos(), f.Name(), "unsupported type %q for field %q", f.Type(), f.Name()))
			fieldErrors = true
		}
	}
	if fieldErrors {
		return optional.None[*schema.Ref]()
	}

	pctx.module.AddData(out)
	return optional.Some[*schema.Ref](dataRef)
}

func ExtractRef(pass *analysis.Pass, pos token.Pos, named *types.Named) (optional.Option[schema.Type], []*schema.Error) {
	var scherrs []*schema.Error
	if named.Obj().Pkg() == nil {
		return optional.None[schema.Type](), scherrs
	}

	// TODO: update visibility is its own analyzer? traversing the schema and updating visibility through references
	//// Update the visibility of the reference if the referencer is exported (ensuring refs are transitively
	//// exported as needed).
	//if isExported {
	//	if decl, ok := pctx.getDeclForTypeName(named.Obj().Name()).Get(); ok {
	//		pctx.markAsExported(decl)
	//	}
	//}

	nodePath := named.Obj().Pkg().Path()
	moduleName, ok := ftlModuleFromGoModule(pass.Pkg.Path()).Get()
	if !ok {
		return optional.None[schema.Type](), append(scherrs,
			noEndColumnErrorf(pass.Fset, pos, "package %q is not in the ftl namespace", named.Obj().Pkg().Path()))
	}

	if !isPathInPkg(pass.Pkg, nodePath) {
		if !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
			return optional.None[schema.Type](), append(scherrs, noEndColumnErrorf(pass.Fset, pos,
				"unsupported external type %q", named.Obj().Pkg().Path()+"."+named.Obj().Name()))
		}
	}
	ref := &schema.Ref{
		Pos:    GoPosToSchemaPos(pass.Fset, pos),
		Module: moduleName,
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
	}
	return optional.Some[schema.Type](ref), scherrs
}

func extractMap(pass *analysis.Pass, pos token.Pos, tnode *types.Map, isExported bool) (optional.Option[schema.Type], []*schema.Error) {
	var scherrs []*schema.Error
	maybeKey, errs := ExtractType(pass, pos, tnode.Key(), isExported)
	scherrs = append(scherrs, errs...)
	key, ok := maybeKey.Get()
	if !ok {
		return optional.None[schema.Type](), scherrs
	}

	maybeValue, errs := ExtractType(pass, pos, tnode.Elem(), isExported)
	scherrs = append(scherrs, errs...)
	if value, ok := maybeValue.Get(); ok {
		return optional.Some[schema.Type](&schema.Map{
			Pos:   GoPosToSchemaPos(pass.Fset, pos),
			Key:   key,
			Value: value,
		}), scherrs
	}

	return optional.None[schema.Type](), scherrs
}

func extractSlice(pass *analysis.Pass, pos token.Pos, tnode *types.Slice, isExported bool) (optional.Option[schema.Type], []*schema.Error) {
	var scherrs []*schema.Error
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return optional.Some[schema.Type](&schema.Bytes{Pos: GoPosToSchemaPos(pass.Fset, pos)}), scherrs
	}
	maybeValue, errs := ExtractType(pass, pos, tnode.Elem(), isExported)
	scherrs = append(scherrs, errs...)

	if value, ok := maybeValue.Get(); ok {
		return optional.Some[schema.Type](&schema.Array{
			Pos:     GoPosToSchemaPos(pass.Fset, pos),
			Element: value,
		}), scherrs
	}

	return optional.None[schema.Type](), scherrs
}

func isPathInPkg(pkg *types.Package, path string) bool {
	if path == pkg.Path() {
		return true
	}
	return strings.HasPrefix(path, pkg.Path()+"/")
}

func ftlModuleFromGoModule(pkgPath string) optional.Option[string] {
	parts := strings.Split(pkgPath, "/")
	if parts[0] != "ftl" {
		return optional.None[string]()
	}
	return optional.Some(strings.TrimSuffix(parts[1], "_test"))
}
