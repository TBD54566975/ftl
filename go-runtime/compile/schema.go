package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/internal/goast"
)

var (
	fset             = token.NewFileSet()
	contextIfaceType = once(func() *types.Interface {
		return mustLoadRef("context", "Context").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})
	errorIFaceType = once(func() *types.Interface {
		return mustLoadRef("builtin", "error").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})

	ftlCallFuncPath   = "github.com/TBD54566975/ftl/go-runtime/ftl.Call"
	ftlConfigFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Config"
	ftlSecretFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Secret" //nolint:gosec
	ftlUnitTypePath   = "github.com/TBD54566975/ftl/go-runtime/ftl.Unit"
	aliasFieldTag     = "json"
)

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Decl]string

type enums map[string]*schema.Enum

func tokenErrorf(pos token.Pos, tokenText string, format string, args ...interface{}) schema.Error {
	goPos := goPosToSchemaPos(pos)
	endColumn := goPos.Column
	if len(tokenText) > 0 {
		endColumn += utf8.RuneCountInString(tokenText)
	}
	return schema.Errorf(goPosToSchemaPos(pos), endColumn, format, args...)
}

func errorf(node ast.Node, format string, args ...interface{}) schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Errorf(pos, endCol, format, args...)
}

func tokenWrapf(pos token.Pos, tokenText string, err error, format string, args ...interface{}) schema.Error {
	goPos := goPosToSchemaPos(pos)
	endColumn := goPos.Column
	if len(tokenText) > 0 {
		endColumn += utf8.RuneCountInString(tokenText)
	}
	return schema.Wrapf(goPos, endColumn, err, format, args...)
}

func wrapf(node ast.Node, err error, format string, args ...interface{}) schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Wrapf(pos, endCol, err, format, args...)
}

// ExtractModuleSchema statically parses Go FTL module source into a schema.Module.
func ExtractModuleSchema(dir string) (NativeNames, *schema.Module, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dir,
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, "./...")
	if err != nil {
		return nil, nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("no packages found in %q, does \"go mod tidy\" need to be run?", dir)
	}
	nativeNames := NativeNames{}
	// Find module name
	module := &schema.Module{}
	merr := []error{}
	for _, pkg := range pkgs {
		moduleName, ok := ftlModuleFromGoModule(pkg.PkgPath).Get()
		if !ok {
			return nil, nil, fmt.Errorf("package %q is not in the ftl namespace", pkg.PkgPath)
		}
		module.Name = moduleName
		if len(pkg.Errors) > 0 {
			for _, perr := range pkg.Errors {
				if len(pkg.Syntax) > 0 {
					merr = append(merr, wrapf(pkg.Syntax[0], perr, "%s", pkg.PkgPath))
				} else {
					merr = append(merr, fmt.Errorf("%s: %w", pkg.PkgPath, perr))
				}
			}
		}
		pctx := &parseContext{pkg: pkg, pkgs: pkgs, module: module, nativeNames: NativeNames{}, enums: enums{}}
		for _, file := range pkg.Syntax {
			err := goast.Visit(file, func(node ast.Node, next func() error) (err error) {
				defer func() {
					if err != nil {
						err = wrapf(node, err, "")
					}
				}()
				switch node := node.(type) {
				case *ast.CallExpr:
					if err := visitCallExpr(pctx, node); err != nil {
						return err
					}

				case *ast.File:
					visitFile(pctx, node)

				case *ast.FuncDecl:
					verb, err := visitFuncDecl(pctx, node)
					if err != nil {
						return err
					}
					pctx.activeVerb = verb
					err = next()
					if err != nil {
						return err
					}
					pctx.activeVerb = nil
					return nil

				case *ast.GenDecl:
					if err = visitGenDecl(pctx, node); err != nil {
						return err
					}

				default:
				}
				return next()
			})
			if err != nil {
				return nil, nil, err
			}
		}
		for decl, nativeName := range pctx.nativeNames {
			nativeNames[decl] = nativeName
		}
		for _, e := range maps.Values(pctx.enums) {
			pctx.module.Decls = append(pctx.module.Decls, e)
		}
	}
	if len(merr) > 0 {
		return nil, nil, errors.Join(merr...)
	}
	return nativeNames, module, schema.ValidateModule(module)
}

func visitCallExpr(pctx *parseContext, node *ast.CallExpr) error {
	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return nil
	}
	switch fn.FullName() {
	case ftlCallFuncPath:
		err := parseCall(pctx, node)
		if err != nil {
			return err
		}

	case ftlConfigFuncPath, ftlSecretFuncPath:
		// Secret/config declaration: ftl.Config[<type>](<name>)
		err := parseConfigDecl(pctx, node, fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseCall(pctx *parseContext, node *ast.CallExpr) error {
	if len(node.Args) != 3 {
		return errorf(node, "call must have exactly three arguments")
	}
	_, verbFn := deref[*types.Func](pctx.pkg, node.Args[1])
	if verbFn == nil {
		if sel, ok := node.Args[1].(*ast.SelectorExpr); ok {
			return errorf(node.Args[1], "call first argument must be a function but is an unresolved reference to %s.%s", sel.X, sel.Sel)
		}
		return errorf(node.Args[1], "call first argument must be a function but is %T", node.Args[1])
	}
	if pctx.activeVerb == nil {
		return nil
	}
	moduleName, ok := ftlModuleFromGoModule(verbFn.Pkg().Path()).Get()
	if !ok {
		return errorf(node.Args[1], "call first argument must be a function in an ftl module")
	}
	ref := &schema.Ref{
		Pos:    goPosToSchemaPos(node.Pos()),
		Module: moduleName,
		Name:   strcase.ToLowerCamel(verbFn.Name()),
	}
	pctx.activeVerb.AddCall(ref)
	return nil
}

func parseConfigDecl(pctx *parseContext, node *ast.CallExpr, fn *types.Func) error {
	var name string
	if len(node.Args) == 1 {
		if literal, ok := node.Args[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
			var err error
			name, err = strconv.Unquote(literal.Value)
			if err != nil {
				return wrapf(node, err, "")
			}
		}
	}
	if name == "" {
		return errorf(node, "config and secret declarations must have a single string literal argument")
	}
	index := node.Fun.(*ast.IndexExpr) //nolint:forcetypeassert

	// Type parameter
	tp := pctx.pkg.TypesInfo.Types[index.Index].Type
	st, err := visitType(pctx, index.Index.Pos(), tp)
	if err != nil {
		return err
	}

	// Add the declaration to the module.
	var decl schema.Decl
	if fn.FullName() == ftlConfigFuncPath {
		decl = &schema.Config{
			Pos:  goPosToSchemaPos(node.Pos()),
			Name: name,
			Type: st,
		}
	} else {
		decl = &schema.Secret{
			Pos:  goPosToSchemaPos(node.Pos()),
			Name: name,
			Type: st,
		}
	}
	pctx.module.Decls = append(pctx.module.Decls, decl)
	return nil
}

func visitFile(pctx *parseContext, node *ast.File) {
	if node.Doc == nil {
		return
	}
	pctx.module.Comments = visitComments(node.Doc)
}

func isType[T types.Type](t types.Type) bool {
	if _, ok := t.(*types.Named); ok {
		t = t.Underlying()
	}
	_, ok := t.(T)
	return ok
}

func checkSignature(sig *types.Signature) (req, resp *types.Var, err error) {
	params := sig.Params()
	results := sig.Results()

	if params.Len() > 2 {
		return nil, nil, fmt.Errorf("must have at most two parameters (context.Context, struct)")
	}
	if params.Len() == 0 {
		return nil, nil, fmt.Errorf("first parameter must be context.Context")
	}
	if !types.AssertableTo(contextIfaceType(), params.At(0).Type()) {
		return nil, nil, fmt.Errorf("first parameter must be of type context.Context but is %s", params.At(0).Type())
	}
	if params.Len() == 2 {
		if !isType[*types.Struct](params.At(1).Type()) {
			return nil, nil, fmt.Errorf("second parameter must be a struct but is %s", params.At(1).Type())
		}
		if params.At(1).Type().String() == ftlUnitTypePath {
			return nil, nil, fmt.Errorf("second parameter must not be ftl.Unit")
		}

		req = params.At(1)
	}

	if results.Len() > 2 {
		return nil, nil, fmt.Errorf("must have at most two results (struct, error)")
	}
	if results.Len() == 0 {
		return nil, nil, fmt.Errorf("must at least return an error")
	}
	if !types.AssertableTo(errorIFaceType(), results.At(results.Len()-1).Type()) {
		return nil, nil, fmt.Errorf("must return an error but is %s", results.At(0).Type())
	}
	if results.Len() == 2 {
		if !isType[*types.Struct](results.At(0).Type()) {
			return nil, nil, fmt.Errorf("first result must be a struct but is %s", results.At(0).Type())
		}
		if results.At(1).Type().String() == ftlUnitTypePath {
			return nil, nil, fmt.Errorf("first result must not be ftl.Unit")
		}
		resp = results.At(0)
	}
	return req, resp, nil
}

func goPosToSchemaPos(pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func goNodePosToSchemaPos(node ast.Node) (schema.Position, int) {
	p := fset.Position(node.Pos())
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}, fset.Position(node.End()).Column
}

func visitGenDecl(pctx *parseContext, node *ast.GenDecl) error {
	switch node.Tok {
	case token.TYPE:
		if node.Doc == nil {
			return nil
		}
		directives, err := parseDirectives(fset, node.Doc)
		if err != nil {
			return err
		}
		for _, dir := range directives {
			switch dir.(type) {
			case *directiveExport:
				if len(node.Specs) != 1 {
					return errorf(node, "error parsing ftl export directive: expected exactly one type "+
						"declaration")
				}
				if pctx.module.Name == "" {
					pctx.module.Name = pctx.pkg.Name
				} else if pctx.module.Name != pctx.pkg.Name {
					return errorf(node, "type export directive must be in the module package")
				}
				if t, ok := node.Specs[0].(*ast.TypeSpec); ok {
					if _, ok := pctx.pkg.TypesInfo.TypeOf(t.Type).Underlying().(*types.Basic); ok {
						enum := &schema.Enum{
							Pos:      goPosToSchemaPos(node.Pos()),
							Comments: visitComments(node.Doc),
						}
						pctx.enums[t.Name.Name] = enum
					}
					err := visitTypeSpec(pctx, t)
					if err != nil {
						return err
					}
				}

			case *directiveIngress:
			}
		}
		return nil

	case token.CONST:
		var typ ast.Expr
		for i, s := range node.Specs {
			v, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}
			// In an iota enum, only the first value has a type.
			// Hydrate this to subsequent values so we can associate them with the enum.
			if i == 0 && isIotaEnum(v) {
				typ = v.Type
			} else if v.Type == nil {
				v.Type = typ
			}
			err := visitValueSpec(pctx, v)
			if err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}

func visitTypeSpec(pctx *parseContext, node *ast.TypeSpec) error {
	if enum, ok := pctx.enums[node.Name.Name]; ok {
		typ, err := visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.TypeOf(node.Type))
		if err != nil {
			return err
		}

		enum.Name = strcase.ToUpperCamel(node.Name.Name)
		enum.Type = typ
		pctx.nativeNames[enum] = node.Name.Name
	} else {
		_, err := visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.Defs[node.Name].Type())
		if err != nil {
			return err
		}
	}
	return nil
}

func visitValueSpec(pctx *parseContext, node *ast.ValueSpec) error {
	var enum *schema.Enum
	if i, ok := node.Type.(*ast.Ident); ok {
		enum = pctx.enums[i.Name]
	}
	if enum == nil {
		return nil
	}
	c, ok := pctx.pkg.TypesInfo.Defs[node.Names[0]].(*types.Const)
	if !ok {
		return errorf(node, "could not extract enum %s: expected exactly one variant name", enum.Name)
	}
	value, err := visitConst(c)
	if err != nil {
		return err
	}
	variant := &schema.EnumVariant{
		Pos:      goPosToSchemaPos(c.Pos()),
		Comments: visitComments(node.Doc),
		Name:     strcase.ToUpperCamel(c.Id()),
		Value:    value,
	}
	enum.Variants = append(enum.Variants, variant)
	return nil
}

func visitFuncDecl(pctx *parseContext, node *ast.FuncDecl) (verb *schema.Verb, err error) {
	if node.Doc == nil {
		return nil, nil
	}
	directives, err := parseDirectives(fset, node.Doc)
	if err != nil {
		return nil, err
	}
	var metadata []schema.Metadata
	isVerb := false
	for _, dir := range directives {
		switch dir := dir.(type) {
		case *directiveExport:
			isVerb = true
			if pctx.module.Name == "" {
				pctx.module.Name = pctx.pkg.Name
			} else if pctx.module.Name != pctx.pkg.Name {
				return nil, errorf(node, "function export directive must be in the module package")
			}

		case *directiveIngress:
			typ := dir.Type
			if typ == "" {
				typ = "http"
			}
			metadata = append(metadata, &schema.MetadataIngress{
				Pos:    dir.Pos,
				Type:   typ,
				Method: dir.Method,
				Path:   dir.Path,
			})
		}
	}
	if !isVerb {
		return nil, nil
	}
	fnt := pctx.pkg.TypesInfo.Defs[node.Name].(*types.Func) //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature)                    //nolint:forcetypeassert
	if sig.Recv() != nil {
		return nil, errorf(node, "ftl:export cannot be a method")
	}
	params := sig.Params()
	results := sig.Results()
	reqt, respt, err := checkSignature(sig)
	if err != nil {
		return nil, wrapf(node, err, "")
	}
	var req schema.Type
	if reqt != nil {
		req, err = visitType(pctx, node.Pos(), params.At(1).Type())
		if err != nil {
			return nil, err
		}
	} else {
		req = &schema.Unit{}
	}
	var resp schema.Type
	if respt != nil {
		resp, err = visitType(pctx, node.Pos(), results.At(0).Type())
		if err != nil {
			return nil, err
		}
	} else {
		resp = &schema.Unit{}
	}
	verb = &schema.Verb{
		Pos:      goPosToSchemaPos(node.Pos()),
		Comments: visitComments(node.Doc),
		Name:     strcase.ToLowerCamel(node.Name.Name),
		Request:  req,
		Response: resp,
		Metadata: metadata,
	}
	pctx.nativeNames[verb] = node.Name.Name
	pctx.module.Decls = append(pctx.module.Decls, verb)
	return verb, nil
}

func parsePathComponents(path string, pos schema.Position) []schema.IngressPathComponent {
	var out []schema.IngressPathComponent
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			out = append(out, &schema.IngressPathParameter{Pos: pos, Name: part[1 : len(part)-1]})
		} else {
			out = append(out, &schema.IngressPathLiteral{Pos: pos, Text: part})
		}
	}
	return out
}

func visitComments(doc *ast.CommentGroup) []string {
	comments := []string{}
	if doc := doc.Text(); doc != "" {
		comments = strings.Split(strings.TrimSpace(doc), "\n")
	}
	return comments
}

func ftlModuleFromGoModule(pkgPath string) optional.Option[string] {
	parts := strings.Split(pkgPath, "/")
	if parts[0] != "ftl" {
		return optional.None[string]()
	}
	return optional.Some(strings.TrimSuffix(parts[1], "_test"))
}

func visitStruct(pctx *parseContext, pos token.Pos, tnode types.Type) (*schema.Ref, error) {
	named, ok := tnode.(*types.Named)
	if !ok {
		return nil, tokenErrorf(pos, tnode.String(), "expected named type but got %s", tnode)
	}
	nodePath := named.Obj().Pkg().Path()
	if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) {
		destModule, ok := ftlModuleFromGoModule(nodePath).Get()
		if !ok {
			return nil, tokenErrorf(pos, nodePath, "struct declared in non-FTL module %s", nodePath)
		}
		dataRef := &schema.Ref{
			Pos:    goPosToSchemaPos(pos),
			Module: destModule,
			Name:   named.Obj().Name(),
		}
		for i := range named.TypeArgs().Len() {
			arg := named.TypeArgs().At(i)
			typeArg, err := visitType(pctx, pos, arg)
			if err != nil {
				return nil, tokenWrapf(pos, arg.String(), err, "type parameter %s", arg.String())
			}

			// Fully qualify the Ref if needed
			if arg, okArg := typeArg.(*schema.Ref); okArg {
				if arg.Module == "" {
					arg.Module = destModule
				}
				typeArg = arg
			}
			dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
		}
		return dataRef, nil
	}

	out := &schema.Data{
		Pos:  goPosToSchemaPos(pos),
		Name: strcase.ToUpperCamel(named.Obj().Name()),
	}
	pctx.nativeNames[out] = named.Obj().Name()
	dataRef := &schema.Ref{
		Pos:    goPosToSchemaPos(pos),
		Module: pctx.module.Name,
		Name:   out.Name,
	}
	for i := range named.TypeParams().Len() {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  goPosToSchemaPos(pos),
			Name: param.Obj().Name(),
		})
		typeArg, err := visitType(pctx, pos, named.TypeArgs().At(i))
		if err != nil {
			return nil, tokenWrapf(pos, param.Obj().Name(), err, "type parameter %s", param.Obj().Name())
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
					out.Comments = visitComments(path.Doc)
				}
			case *ast.GenDecl:
				if path.Doc != nil {
					out.Comments = visitComments(path.Doc)
				}
			}
		}
	}

	s, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, tokenErrorf(pos, named.String(), "expected struct but got %s", named)
	}
	for i := range s.NumFields() {
		f := s.Field(i)
		ft, err := visitType(pctx, f.Pos(), f.Type())
		if err != nil {
			return nil, tokenWrapf(f.Pos(), f.Name(), err, "field %s", f.Name())
		}

		// Check if field is exported
		if len(f.Name()) > 0 && unicode.IsLower(rune(f.Name()[0])) {
			return nil, tokenErrorf(f.Pos(), f.Name(), "params field %s must be exported by starting with an uppercase letter", f.Name())
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
	}
	pctx.module.AddData(out)
	return dataRef, nil
}

func visitConst(cnode *types.Const) (schema.Value, error) {
	if b, ok := cnode.Type().Underlying().(*types.Basic); ok {
		switch b.Kind() {
		case types.String:
			value, err := strconv.Unquote(cnode.Val().String())
			if err != nil {
				return nil, err
			}
			return &schema.StringValue{Pos: goPosToSchemaPos(cnode.Pos()), Value: value}, nil

		case types.Int, types.Int64:
			value, err := strconv.ParseInt(cnode.Val().String(), 10, 64)
			if err != nil {
				return nil, err
			}
			return &schema.IntValue{Pos: goPosToSchemaPos(cnode.Pos()), Value: int(value)}, nil
		default:
			return nil, tokenErrorf(cnode.Pos(), b.Name(), "unsupported basic type %s", b)
		}
	}
	return nil, tokenErrorf(cnode.Pos(), cnode.Type().String(), "unsupported const type %s", cnode.Type())
}

func visitType(pctx *parseContext, pos token.Pos, tnode types.Type) (schema.Type, error) {
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return &schema.Ref{Pos: goPosToSchemaPos(pos), Name: tparam.Obj().Id()}, nil
	}
	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			nodePath := named.Obj().Pkg().Path()
			if pctx.enums[named.Obj().Name()] != nil {
				return &schema.Ref{
					Pos:    goPosToSchemaPos(pos),
					Module: pctx.module.Name,
					Name:   named.Obj().Name(),
				}, nil
			} else if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) {
				// If this type is named and declared in another module, it's a reference.
				// The only basic-typed references supported are enums.
				if !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
					return nil, fmt.Errorf("unsupported external type %s", named.Obj().Pkg().Path()+"."+named.Obj().Name())
				}
				base := path.Dir(pctx.pkg.PkgPath)
				destModule := path.Base(strings.TrimPrefix(nodePath, base+"/"))
				enumRef := &schema.Ref{
					Pos:    goPosToSchemaPos(pos),
					Module: destModule,
					Name:   named.Obj().Name(),
				}
				return enumRef, nil
			}
		}

		switch underlying.Kind() {
		case types.String:
			return &schema.String{Pos: goPosToSchemaPos(pos)}, nil

		case types.Int, types.Int64:
			return &schema.Int{Pos: goPosToSchemaPos(pos)}, nil

		case types.Bool:
			return &schema.Bool{Pos: goPosToSchemaPos(pos)}, nil

		case types.Float64:
			return &schema.Float{Pos: goPosToSchemaPos(pos)}, nil

		default:
			return nil, tokenErrorf(pos, underlying.Name(), "unsupported basic type %s", underlying)
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			return visitStruct(pctx, pos, tnode)
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return &schema.Time{Pos: goPosToSchemaPos(pos)}, nil

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Unit":
			return &schema.Unit{Pos: goPosToSchemaPos(pos)}, nil

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Option":
			underlying, err := visitType(pctx, pos, named.TypeArgs().At(0))
			if err != nil {
				return nil, err
			}
			return &schema.Optional{Type: underlying}, nil
		default:
			nodePath := named.Obj().Pkg().Path()
			if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) && !strings.HasPrefix(nodePath, "ftl/") {
				return nil, fmt.Errorf("unsupported external type %s", nodePath+"."+named.Obj().Name())
			}
			return visitStruct(pctx, pos, tnode)
		}

	case *types.Map:
		return visitMap(pctx, pos, underlying)

	case *types.Slice:
		return visitSlice(pctx, pos, underlying)

	case *types.Interface:
		if underlying.String() == "any" {
			return &schema.Any{Pos: goPosToSchemaPos(pos)}, nil
		}
		return nil, tokenErrorf(pos, "", "unsupported type %q", tnode)

	default:
		return nil, tokenErrorf(pos, "", "unsupported type %q", tnode)
	}
}

func visitMap(pctx *parseContext, pos token.Pos, tnode *types.Map) (*schema.Map, error) {
	key, err := visitType(pctx, pos, tnode.Key())
	if err != nil {
		return nil, err
	}
	value, err := visitType(pctx, pos, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Map{
		Pos:   goPosToSchemaPos(pos),
		Key:   key,
		Value: value,
	}, nil
}

func visitSlice(pctx *parseContext, pos token.Pos, tnode *types.Slice) (schema.Type, error) {
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return &schema.Bytes{Pos: goPosToSchemaPos(pos)}, nil
	}
	value, err := visitType(pctx, pos, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Array{
		Pos:     goPosToSchemaPos(pos),
		Element: value,
	}, nil
}

func once[T any](f func() T) func() T {
	var once sync.Once
	var t T
	return func() T {
		once.Do(func() { t = f() })
		return t
	}
}

// Lazy load the compile-time reference from a package.
func mustLoadRef(pkg, name string) types.Object {
	pkgs, err := packages.Load(&packages.Config{Fset: fset, Mode: packages.NeedTypes}, pkg)
	if err != nil {
		panic(err)
	}
	if len(pkgs) != 1 {
		panic("expected one package")
	}
	obj := pkgs[0].Types.Scope().Lookup(name)
	if obj == nil {
		panic("interface not found")
	}
	return obj
}

func deref[T types.Object](pkg *packages.Package, node ast.Expr) (string, T) {
	var obj T
	switch node := node.(type) {
	case *ast.Ident:
		obj, _ = pkg.TypesInfo.Uses[node].(T)
		return "", obj

	case *ast.SelectorExpr:
		x, ok := node.X.(*ast.Ident)
		if !ok {
			return "", obj
		}
		obj, _ = pkg.TypesInfo.Uses[node.Sel].(T)
		return x.Name, obj

	case *ast.IndexExpr:
		return deref[T](pkg, node.X)

	default:
		return "", obj
	}
}

type parseContext struct {
	pkg         *packages.Package
	pkgs        []*packages.Package
	module      *schema.Module
	nativeNames NativeNames
	enums       enums
	activeVerb  *schema.Verb
}

// pathEnclosingInterval returns the PackageInfo and ast.Node that
// contain source interval [start, end), and all the node's ancestors
// up to the AST root.  It searches all ast.Files of all packages in prog.
// exact is defined as for astutil.PathEnclosingInterval.
//
// The zero value is returned if not found.
func (p *parseContext) pathEnclosingInterval(start, end token.Pos) (pkg *packages.Package, path []ast.Node, exact bool) {
	for _, info := range p.pkgs {
		for _, f := range info.Syntax {
			if f.Pos() == token.NoPos {
				// This can happen if the parser saw
				// too many errors and bailed out.
				// (Use parser.AllErrors to prevent that.)
				continue
			}
			if !tokenFileContainsPos(fset.File(f.Pos()), start) {
				continue
			}
			if path, exact := astutil.PathEnclosingInterval(f, start, end); path != nil {
				return info, path, exact
			}
		}
	}
	return nil, nil, false
}

func tokenFileContainsPos(f *token.File, pos token.Pos) bool {
	p := int(pos)
	base := f.Base()
	return base <= p && p < base+f.Size()
}

func isIotaEnum(node ast.Node) bool {
	switch t := node.(type) {
	case *ast.ValueSpec:
		if len(t.Values) != 1 {
			return false
		}
		return isIotaEnum(t.Values[0])
	case *ast.Ident:
		return t.Name == "iota"
	case *ast.BinaryExpr:
		return isIotaEnum(t.X) || isIotaEnum(t.Y)
	default:
		return false
	}
}
