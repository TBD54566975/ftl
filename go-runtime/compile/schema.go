package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/backend/common/goast"
	"github.com/TBD54566975/ftl/backend/schema"
)

var (
	fset             = token.NewFileSet()
	contextIfaceType = once(func() *types.Interface {
		return mustLoadRef("context", "Context").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})
	errorIFaceType = once(func() *types.Interface {
		return mustLoadRef("builtin", "error").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})
	ftlCallFuncPath = "github.com/TBD54566975/ftl/go-runtime/sdk.Call"
)

// ExtractModuleSchema statically parses Go FTL module source into a schema.Module.
func ExtractModuleSchema(dir string) (*schema.Module, error) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dir,
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, "./...")
	if err != nil {
		return &schema.Module{}, err
	}
	if len(pkgs) == 0 {
		return &schema.Module{}, fmt.Errorf("no packages found in %q, does \"go mod tidy\" need to be run?", dir)
	}
	module := &schema.Module{}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("%s: %w", pkg.PkgPath, pkg.Errors[0])
		}
		pctx := &parseContext{pkg: pkg, pkgs: pkgs, module: module}
		for _, file := range pkg.Syntax {
			var verb *schema.Verb
			err := goast.Visit(file, func(node ast.Node, next func() error) (err error) {
				defer func() {
					if err != nil {
						err = fmt.Errorf("%s: %w", fset.Position(node.Pos()).String(), err)
					}
				}()
				switch node := node.(type) {
				case *ast.CallExpr:
					// Only track calls when we're in a verb.
					if verb != nil {
						if err := visitCallExpr(pctx, verb, node); err != nil {
							return err
						}
					}

				case *ast.File:
					if err := visitFile(pctx, node); err != nil {
						return err
					}

				case *ast.FuncDecl:
					verb, err = visitFuncDecl(pctx, node)
					if err != nil {
						return err
					}
					err = next()
					if err != nil {
						return err
					}
					verb = nil
					return nil

				case nil:
				default:
				}
				return next()
			})
			if err != nil {
				return nil, err
			}
		}
	}
	if module.Name == "" {
		return module, fmt.Errorf("//ftl:module directive is required")
	}
	return module, schema.ValidateModule(module)
}

func visitCallExpr(pctx *parseContext, verb *schema.Verb, node *ast.CallExpr) error {
	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return nil
	}
	if fn.FullName() != ftlCallFuncPath {
		return nil
	}
	if len(node.Args) != 3 {
		return errors.New("call must have exactly three arguments")
	}
	_, verbFn := deref[*types.Func](pctx.pkg, node.Args[1])
	if verbFn == nil {
		return fmt.Errorf("call first argument must be a function but is %s", node.Args[1])
	}
	moduleName := verbFn.Pkg().Name()
	if moduleName == pctx.pkg.Name {
		moduleName = ""
	}
	ref := &schema.VerbRef{
		Pos:    goPosToSchemaPos(node.Pos()),
		Module: moduleName,
		Name:   strcase.ToLowerCamel(verbFn.Name()),
	}
	verb.AddCall(ref)
	return nil
}

func visitFile(pctx *parseContext, node *ast.File) error {
	if node.Doc == nil {
		return nil
	}
	directives, err := parseDirectives(fset, node.Doc)
	if err != nil {
		return err
	}
	pctx.module.Comments = visitComments(node.Doc)
	for _, dir := range directives {
		switch dir := dir.(type) {
		case *directiveModule:
			if dir.Name != pctx.pkg.Name {
				return fmt.Errorf("%s: FTL module name %q does not match Go package name %q", dir, dir.Name, pctx.pkg.Name)
			}
			pctx.module.Name = dir.Name

		default:
			return fmt.Errorf("%s: invalid directive", dir)
		}
	}
	return nil
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
		resp = results.At(0)
	}
	if params.Len() == 1 && results.Len() == 1 {
		return nil, nil, fmt.Errorf("must either accept an input or return a result, but does neither")
	}
	return req, resp, nil
}

func goPosToSchemaPos(pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
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
		case *directiveModule:

		case *directiveVerb:
			isVerb = true

		case *directiveIngress:
			metadata = append(metadata, &schema.MetadataIngress{
				Pos:    dir.Pos,
				Type:   dir.Type,
				Method: dir.Method,
				Path:   dir.Path,
			})

		default:
			panic(fmt.Sprintf("unsupported directive %T", dir))
		}
	}
	if !isVerb {
		return nil, nil
	}
	fnt := pctx.pkg.TypesInfo.Defs[node.Name].(*types.Func) //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature)                    //nolint:forcetypeassert
	if sig.Recv() != nil {
		return nil, fmt.Errorf("ftl:verb cannot be a method")
	}
	params := sig.Params()
	results := sig.Results()
	reqt, respt, err := checkSignature(sig)
	if err != nil {
		return nil, err
	}
	var req schema.Type
	if reqt != nil {
		req, err = visitType(pctx, node, params.At(1).Type())
		if err != nil {
			return nil, err
		}
	} else {
		req = &schema.Unit{}
	}
	var resp schema.Type
	if respt != nil {
		resp, err = visitType(pctx, node, results.At(0).Type())
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

func visitStruct(pctx *parseContext, node ast.Node, tnode types.Type) (*schema.DataRef, error) {
	named, ok := tnode.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("expected named type but got %s", tnode)
	}
	nodePath := named.Obj().Pkg().Path()
	if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) {
		base := path.Dir(pctx.pkg.PkgPath)
		destModule := path.Base(strings.TrimPrefix(nodePath, base+"/"))
		return &schema.DataRef{
			Pos:    goPosToSchemaPos(node.Pos()),
			Module: destModule,
			Name:   named.Obj().Name(),
		}, nil
	}
	out := &schema.Data{
		Pos:  goPosToSchemaPos(node.Pos()),
		Name: named.Obj().Name(),
	}
	dataRef := &schema.DataRef{
		Pos:  goPosToSchemaPos(node.Pos()),
		Name: out.Name,
	}
	for i := 0; i < named.TypeParams().Len(); i++ {
		param := named.TypeParams().At(i)
		out.TypeParameters = append(out.TypeParameters, &schema.TypeParameter{
			Pos:  goPosToSchemaPos(node.Pos()),
			Name: param.Obj().Name(),
		})
		typeArg, err := visitType(pctx, node, named.TypeArgs().At(i))
		if err != nil {
			return nil, fmt.Errorf("type parameter %s: %w", param.Obj().Name(), err)
		}
		dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
	}

	// If the struct is generic, we need to use the origin type to get the
	// fields.
	if named.TypeParams().Len() > 0 {
		named = named.Origin()
	}

	// Find type declaration so we can extract comments.
	pos := named.Obj().Pos()
	pkg, path, _ := pctx.pathEnclosingInterval(pos, pos)
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
		return nil, fmt.Errorf("expected struct but got %s", named)
	}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		ft, err := visitType(pctx, node, f.Type())
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", f.Name(), err)
		}
		out.Fields = append(out.Fields, &schema.Field{
			Pos:  goPosToSchemaPos(node.Pos()),
			Name: strcase.ToLowerCamel(f.Name()),
			Type: ft,
		})
	}
	pctx.module.AddData(out)
	return dataRef, nil
}

func visitType(pctx *parseContext, node ast.Node, tnode types.Type) (schema.Type, error) {
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return &schema.TypeParameter{Name: tparam.Obj().Id()}, nil
	}
	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		switch underlying.Kind() {
		case types.String:
			return &schema.String{Pos: goPosToSchemaPos(node.Pos())}, nil

		case types.Int, types.Int64:
			return &schema.Int{Pos: goPosToSchemaPos(node.Pos())}, nil

		case types.Bool:
			return &schema.Bool{Pos: goPosToSchemaPos(node.Pos())}, nil

		case types.Float64:
			return &schema.Float{Pos: goPosToSchemaPos(node.Pos())}, nil

		default:
			return nil, fmt.Errorf("unsupported basic type %s", underlying)
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			return visitStruct(pctx, node, tnode)
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return &schema.Time{Pos: goPosToSchemaPos(node.Pos())}, nil

		case "github.com/TBD54566975/ftl/go-runtime/sdk.Unit":
			return &schema.Unit{Pos: goPosToSchemaPos(node.Pos())}, nil

		case "github.com/TBD54566975/ftl/go-runtime/sdk.Option":
			underlying, err := visitType(pctx, node, named.TypeArgs().At(0))
			if err != nil {
				return nil, err
			}
			return &schema.Optional{Type: underlying}, nil

		default:
			return visitStruct(pctx, node, tnode)
		}

	case *types.Map:
		return visitMap(pctx, node, underlying)

	case *types.Slice:
		return visitSlice(pctx, node, underlying)

	case *types.Interface:
		if underlying.String() == "any" {
			return &schema.Any{Pos: goPosToSchemaPos(node.Pos())}, nil
		}
		return nil, fmt.Errorf("%s: unsupported type %T", goPosToSchemaPos(node.Pos()), node)

	default:
		return nil, fmt.Errorf("%s: unsupported type %T", goPosToSchemaPos(node.Pos()), node)
	}
}

func visitMap(pctx *parseContext, node ast.Node, tnode *types.Map) (*schema.Map, error) {
	key, err := visitType(pctx, node, tnode.Key())
	if err != nil {
		return nil, err
	}
	value, err := visitType(pctx, node, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Map{
		Pos:   goPosToSchemaPos(node.Pos()),
		Key:   key,
		Value: value,
	}, nil
}

func visitSlice(pctx *parseContext, node ast.Node, tnode *types.Slice) (schema.Type, error) {
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return &schema.Bytes{Pos: goPosToSchemaPos(node.Pos())}, nil
	}
	value, err := visitType(pctx, node, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Array{
		Pos:     goPosToSchemaPos(node.Pos()),
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

	default:
		return "", obj
	}
}

type parseContext struct {
	pkg    *packages.Package
	pkgs   []*packages.Package
	module *schema.Module
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
