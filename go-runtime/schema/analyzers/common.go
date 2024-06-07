package analyzers

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/ast/astutil"
	"github.com/alecthomas/types/optional"
)

type DiagnosticCategory string

const (
	Info  DiagnosticCategory = "info"
	Warn  DiagnosticCategory = "warn"
	Error DiagnosticCategory = "error"
)

func (e DiagnosticCategory) ToErrorLevel() schema.ErrorLevel {
	switch e {
	case Info:
		return schema.INFO
	case Warn:
		return schema.WARN
	case Error:
		return schema.ERROR
	default:
		panic(fmt.Sprintf("unknown diagnostic category %q", e))
	}
}

type NativeNames map[schema.Node]string

var (
	aliasFieldTag = "json"
)

// TODO: maybe don't need NativeNames from extractors once we process refs/native names as an initial analyzer?
type result struct {
	decls       []schema.Decl
	nativeNames NativeNames
}

func extractComments(doc *ast.CommentGroup) []string {
	comments := []string{}
	if doc := doc.Text(); doc != "" {
		comments = strings.Split(strings.TrimSpace(doc), "\n")
	}
	return comments
}

func extractType(pass *analysis.Pass, pos token.Pos, tnode types.Type, isExported bool) optional.Option[schema.Type] {
	fset := pass.Fset
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return optional.Some[schema.Type](&schema.Ref{Pos: goPosToSchemaPos(fset, pos), Name: tparam.Obj().Id()})
	}

	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			return extractRef(pass, pos, named, isExported)
		}
		switch underlying.Kind() {
		case types.String:
			return optional.Some[schema.Type](&schema.String{Pos: goPosToSchemaPos(fset, pos)})

		case types.Int, types.Int64:
			return optional.Some[schema.Type](&schema.Int{Pos: goPosToSchemaPos(fset, pos)})

		case types.Bool:
			return optional.Some[schema.Type](&schema.Bool{Pos: goPosToSchemaPos(fset, pos)})

		case types.Float64:
			return optional.Some[schema.Type](&schema.Float{Pos: goPosToSchemaPos(fset, pos)})

		default:
			return optional.None[schema.Type]()
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			pass.Report(noEndColumnErrorf(fset, pos, "expected named type but got %s", tnode))
			return optional.None[schema.Type]()
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return optional.Some[schema.Type](&schema.Time{Pos: goPosToSchemaPos(fset, pos)})

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Unit":
			return optional.Some[schema.Type](&schema.Unit{Pos: goPosToSchemaPos(fset, pos)})

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Option":
			typ := extractType(pass, pos, named.TypeArgs().At(0), isExported)
			if underlying, ok := typ.Get(); ok {
				return optional.Some[schema.Type](&schema.Optional{Pos: goPosToSchemaPos(pass.Fset, pos), Type: underlying})
			}
			return optional.None[schema.Type]()

		default:
			nodePath := named.Obj().Pkg().Path()
			if !isPathInPkg(pass.Pkg, nodePath) && !strings.HasPrefix(nodePath, "ftl/") {
				pass.Report(noEndColumnErrorf(fset, pos, "unsupported external type %s", nodePath+"."+named.Obj().Name()))
				return optional.None[schema.Type]()
			}
			return extractData(pass, pos, tnode, isExported)
		}

	case *types.Map:
		return extractMap(pass, pos, underlying, isExported)

	case *types.Slice:
		return extractSlice(pass, pos, underlying, isExported)

	case *types.Interface:
		if underlying.String() == "any" {
			return optional.Some[schema.Type](&schema.Any{Pos: goPosToSchemaPos(fset, pos)})
		}
		if named, ok := tnode.(*types.Named); ok {
			return extractRef(pass, pos, named, isExported)
		}
		return optional.None[schema.Type]()

	default:
		return optional.None[schema.Type]()
	}
}

func extractRef(pass *analysis.Pass, pos token.Pos, named *types.Named, isExported bool) optional.Option[schema.Type] {
	if named.Obj().Pkg() == nil {
		return optional.None[schema.Type]()
	}

	nodePath := named.Obj().Pkg().Path()
	moduleName, err := ftlModuleFromGoModule(nodePath)
	if err != nil {
		pass.Report(noEndColumnWrapf(pass.Fset, pos, err, ""))
		return optional.None[schema.Type]()
	}

	if !isPathInPkg(pass.Pkg, nodePath) {
		if !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
			pass.Report(noEndColumnErrorf(pass.Fset, pos, "unsupported external type %q", named.Obj().Pkg().Path()+"."+named.Obj().Name()))
			return optional.None[schema.Type]()
		}
	}

	ref := &schema.Ref{
		Pos:    goPosToSchemaPos(pass.Fset, pos),
		Module: moduleName,
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
	}
	for i := range named.TypeArgs().Len() {
		typeArg, ok := extractType(pass, pos, named.TypeArgs().At(i), isExported).Get()
		if !ok {
			pass.Report(tokenErrorf(pass.Fset, pos, named.TypeArgs().At(i).String(),
				"unsupported type %q for type argument", named.TypeArgs().At(i)))
			continue
		}

		// Fully qualify the Ref if needed
		if r, okArg := typeArg.(*schema.Ref); okArg {
			if r.Module == "" {
				r.Module = moduleName
			}
			typeArg = r
		}
		ref.TypeParameters = append(ref.TypeParameters, typeArg)
	}

	return optional.Some[schema.Type](ref)
}

// TODO: probably don't need this in common and can move to data extractor once implemented
func extractData(pass *analysis.Pass, pos token.Pos, tnode types.Type, isExported bool) optional.Option[schema.Type] {
	fset := pass.Fset
	named, ok := tnode.(*types.Named)
	if !ok {
		pass.Report(noEndColumnErrorf(fset, pos, "expected named type but got %s", tnode))
		return optional.None[schema.Type]()
	}

	nodePath := named.Obj().Pkg().Path()
	nodeModule, err := ftlModuleFromGoModule(nodePath)
	if err != nil {
		pass.Report(noEndColumnWrapf(fset, pos, err, ""))
		return optional.None[schema.Type]()
	}
	if !isPathInPkg(pass.Pkg, nodePath) {
		return extractRef(pass, pos, named, isExported)
	}

	out := &schema.Data{
		Pos:    goPosToSchemaPos(fset, pos),
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
		Export: isExported,
	}
	// ectx.addNativeName(out, named.Obj().Name()) <-- TODO: add back when data extractor is implemented
	dataRef := &schema.Ref{
		Pos:    goPosToSchemaPos(fset, pos),
		Module: nodeModule,
		Name:   out.Name,
	}
	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  goPosToSchemaPos(fset, pos),
			Name: param.Obj().Name(),
		})
		typeArgs := named.TypeArgs()
		if typeArgs == nil {
			continue
		}
		typeArg, ok := extractType(pass, pos, typeArgs.At(i), isExported).Get()
		if !ok {
			pass.Report(tokenErrorf(fset, pos, typeArgs.At(i).String(),
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
	maybePath, _ := pathEnclosingInterval(pass, namedPos, namedPos)
	if path, ok := maybePath.Get(); ok {
		for i := len(path) - 1; i >= 0; i-- {
			// We have to check both the type spec and the gen decl because the
			// type could be declared as either "type Foo struct { ... }" or
			// "type ( Foo struct { ... } )"
			switch path := path[i].(type) {
			case *ast.TypeSpec:
				if path.Doc != nil {
					out.Comments = extractComments(path.Doc)
				}
			case *ast.GenDecl:
				if path.Doc != nil {
					out.Comments = extractComments(path.Doc)
				}
			}
		}
	}

	s, ok := named.Underlying().(*types.Struct)
	if !ok {
		pass.Report(noEndColumnErrorf(fset, pos, "expected struct but got %s", named))
		return optional.None[schema.Type]()
	}

	fieldErrors := false
	for i := range s.NumFields() {
		f := s.Field(i)
		if ft, ok := extractType(pass, f.Pos(), f.Type(), isExported).Get(); ok {
			// Check if field is exported
			if len(f.Name()) > 0 && unicode.IsLower(rune(f.Name()[0])) {
				pass.Report(tokenErrorf(fset, f.Pos(), f.Name(),
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
					Pos:   goPosToSchemaPos(pass.Fset, pos),
					Kind:  schema.AliasKindJSON,
					Alias: jsonFieldName,
				})
			}
			out.Fields = append(out.Fields, &schema.Field{
				Pos:      goPosToSchemaPos(pass.Fset, pos),
				Name:     strcase.ToLowerCamel(f.Name()),
				Type:     ft,
				Metadata: metadata,
			})
		} else {
			pass.Report(tokenErrorf(fset, f.Pos(), f.Name(), "unsupported type %q for field %q", f.Type(), f.Name()))
			fieldErrors = true
		}
	}
	if fieldErrors {
		return optional.None[schema.Type]()
	}

	// ectx.module.AddData(out) <--  TODO: add back when data extractor is implemented
	return optional.Some[schema.Type](dataRef)
}

func extractMap(pass *analysis.Pass, pos token.Pos, tnode *types.Map, isExported bool) optional.Option[schema.Type] {
	key, ok := extractType(pass, pos, tnode.Key(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	value, ok := extractType(pass, pos, tnode.Elem(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Map{Pos: goPosToSchemaPos(pass.Fset, pos), Key: key, Value: value})
}

func extractSlice(pass *analysis.Pass, pos token.Pos, tnode *types.Slice, isExported bool) optional.Option[schema.Type] {
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return optional.Some[schema.Type](&schema.Bytes{Pos: goPosToSchemaPos(pass.Fset, pos)})
	}

	value, ok := extractType(pass, pos, tnode.Elem(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Array{
		Pos:     goPosToSchemaPos(pass.Fset, pos),
		Element: value,
	})
}

func goPosToSchemaPos(fset *token.FileSet, pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func tokenFileContainsPos(f *token.File, pos token.Pos) bool {
	p := int(pos)
	base := f.Base()
	return base <= p && p < base+f.Size()
}

func isPathInPkg(pkg *types.Package, path string) bool {
	if path == pkg.Path() {
		return true
	}
	return strings.HasPrefix(path, pkg.Path()+"/")
}

// pathEnclosingInterval returns the PackageInfo and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches all ast.Files of all packages in prog.
// exact is defined as for astutil.PathEnclosingInterval.
//
// An empty path optional is returned if not found.
func pathEnclosingInterval(pass *analysis.Pass, start, end token.Pos) (path optional.Option[[]ast.Node], exact bool) {
	for _, f := range pass.Files {
		if f.Pos() == token.NoPos {
			// This can happen if the parser saw
			// too many errors and bailed out.
			// (Use parser.AllErrors to prevent that.)
			continue
		}
		if !tokenFileContainsPos(pass.Fset.File(f.Pos()), start) {
			continue
		}
		if path, exact := astutil.PathEnclosingInterval(f, start, end); path != nil {
			return optional.Some(path), exact
		}
	}

	return optional.None[[]ast.Node](), false
}

func ftlModuleFromGoModule(pkgPath string) (string, error) {
	parts := strings.Split(pkgPath, "/")
	if parts[0] != "ftl" {
		return "", fmt.Errorf("package %q is not in the ftl namespace", pkgPath)
	}
	return strings.TrimSuffix(parts[1], "_test"), nil
}

func errorf(node ast.Node, format string, args ...interface{}) analysis.Diagnostic {
	return errorfAtPos(node.Pos(), node.End(), format, args...)
}

func errorfAtPos(pos token.Pos, end token.Pos, format string, args ...interface{}) analysis.Diagnostic {
	return analysis.Diagnostic{Pos: pos, End: end, Message: fmt.Sprintf(format, args...), Category: string(Error)}
}

func noEndColumnErrorf(fset *token.FileSet, pos token.Pos, format string, args ...interface{}) analysis.Diagnostic {
	return tokenErrorf(fset, pos, "", format, args...)
}

func tokenErrorf(fset *token.FileSet, pos token.Pos, tokenText string, format string, args ...interface{}) analysis.Diagnostic {
	endCol := fset.Position(pos).Column
	if len(tokenText) > 0 {
		endCol += utf8.RuneCountInString(tokenText)
	}
	return errorfAtPos(pos, fset.File(pos).Pos(endCol), format, args...)
}

func noEndColumnWrapf(fset *token.FileSet, pos token.Pos, err error, format string, args ...interface{}) analysis.Diagnostic {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	args = append(args, err)
	return tokenErrorf(fset, pos, "", format, args...)
}

func wrapf(node ast.Node, err error, format string, args ...interface{}) analysis.Diagnostic {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	args = append(args, err)
	return errorfAtPos(node.Pos(), node.End(), format, args...)
}
