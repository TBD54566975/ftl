package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/participle/v2"
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
	pctx.module.Comments = parseComments(node.Doc)
	for _, dir := range directives {
		switch dir.kind {
		case "module":
			moduleAttr := dir.getPositional(0)
			if moduleAttr == nil {
				return fmt.Errorf("%s: module not specified", dir)
			}
			moduleName := moduleAttr.Value.Text()
			if moduleAttr.Value.Text() != pctx.pkg.Name {
				return fmt.Errorf("%s: FTL module name %q does not match Go package name %q", dir, moduleName, pctx.pkg.Name)
			}
			pctx.module.Name = moduleName

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

func checkSignature(sig *types.Signature) error {
	params := sig.Params()
	results := sig.Results()
	if params.Len() != 2 {
		return fmt.Errorf("must have exactly two parameters in the form (context.Context, struct) but has %d", params.Len())
	}
	if results.Len() != 2 {
		return fmt.Errorf("must have exactly two result values in the form (error, struct) but has %d", results.Len())
	}
	if !types.AssertableTo(contextIfaceType(), params.At(0).Type()) {
		return fmt.Errorf("first parameter must be of type context.Context but is %s", params.At(0).Type())
	}
	if !isType[*types.Struct](params.At(1).Type()) {
		return fmt.Errorf("second parameter must be a struct but is %s", params.At(1).Type())
	}
	if !types.AssertableTo(errorIFaceType(), results.At(1).Type()) {
		return fmt.Errorf("first result must be an error but is %s", results.At(0).Type())
	}
	if !isType[*types.Struct](results.At(0).Type()) {
		return fmt.Errorf("first result must be a struct but is %s", results.At(0).Type())
	}
	return nil
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
		switch dir.kind {
		case "verb":
			isVerb = true
		case "ingress":
			methodAttr := dir.getPositional(0)
			if methodAttr == nil {
				return nil, fmt.Errorf("%s: ingress method not specified", dir.pos)
			}
			pathAttr := dir.getPositional(1)
			if pathAttr == nil {
				return nil, fmt.Errorf("%s: ingress path not specified", dir.pos)
			}
			metadata = append(metadata, &schema.MetadataIngress{
				Pos:    dir.pos,
				Method: methodAttr.Value.Text(),
				Path:   pathAttr.Value.Text(),
			})
		default:
			return nil, fmt.Errorf("invalid directive: %s", dir)
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
	if err := checkSignature(sig); err != nil {
		return nil, err
	}
	req, err := parseStruct(pctx, node, params.At(1).Type())
	if err != nil {
		return nil, err
	}
	resp, err := parseStruct(pctx, node, results.At(0).Type())
	if err != nil {
		return nil, err
	}
	verb = &schema.Verb{
		Pos:      goPosToSchemaPos(node.Pos()),
		Comments: parseComments(node.Doc),
		Name:     strcase.ToLowerCamel(node.Name.Name),
		Request:  req,
		Response: resp,
		Metadata: metadata,
	}
	pctx.module.Decls = append(pctx.module.Decls, verb)
	return verb, nil
}

func parseComments(doc *ast.CommentGroup) []string {
	comments := []string{}
	if doc := doc.Text(); doc != "" {
		comments = strings.Split(strings.TrimSpace(doc), "\n")
	}
	return comments
}

func parseStruct(pctx *parseContext, node ast.Node, tnode types.Type) (*schema.DataRef, error) {
	named, ok := tnode.(*types.Named)
	if !ok {
		return nil, fmt.Errorf("expected named type but got %s", tnode)
	}
	out := &schema.Data{
		Pos:  goPosToSchemaPos(node.Pos()),
		Name: named.Obj().Name(),
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
					out.Comments = parseComments(path.Doc)
				}
			case *ast.GenDecl:
				if path.Doc != nil {
					out.Comments = parseComments(path.Doc)
				}
			}
		}
	}

	s, ok := tnode.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct but got %s", tnode)
	}
	for i := 0; i < s.NumFields(); i++ {
		f := s.Field(i)
		ft, err := parseType(pctx, node, f.Type())
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
	return &schema.DataRef{
		Pos:  goPosToSchemaPos(node.Pos()),
		Name: out.Name,
	}, nil
}

func parseType(pctx *parseContext, node ast.Node, tnode types.Type) (schema.Type, error) {
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
			return parseStruct(pctx, node, tnode)
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return &schema.Time{Pos: goPosToSchemaPos(node.Pos())}, nil

		case "github.com/TBD54566975/ftl/go-runtime/sdk.Option":
			underlying, err := parseType(pctx, node, named.TypeArgs().At(0))
			if err != nil {
				return nil, err
			}
			return &schema.Optional{Type: underlying}, nil

		default:
			return parseStruct(pctx, node, tnode)
		}

	case *types.Map:
		return parseMap(pctx, node, underlying)

	case *types.Slice:
		return parseSlice(pctx, node, underlying)

	default:
		return nil, fmt.Errorf("%s: unsupported type %T", goPosToSchemaPos(node.Pos()), node)
	}
}

func parseMap(pctx *parseContext, node ast.Node, tnode *types.Map) (*schema.Map, error) {
	key, err := parseType(pctx, node, tnode.Key())
	if err != nil {
		return nil, err
	}
	value, err := parseType(pctx, node, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Map{
		Pos:   goPosToSchemaPos(node.Pos()),
		Key:   key,
		Value: value,
	}, nil
}

func parseSlice(pctx *parseContext, node ast.Node, tnode *types.Slice) (*schema.Array, error) {
	value, err := parseType(pctx, node, tnode.Elem())
	if err != nil {
		return nil, err
	}
	return &schema.Array{
		Pos:     goPosToSchemaPos(node.Pos()),
		Element: value,
	}, nil
}

type ftlDirective struct {
	kind  string
	attrs []directiveAttr
	pos   schema.Position
}

func (f ftlDirective) String() string {
	out := &strings.Builder{}
	fmt.Fprintf(out, "//ftl:%s", f.kind)
	for _, attr := range f.attrs {
		if attr.Key != nil {
			fmt.Fprintf(out, " %s=", *attr.Key)
		} else {
			fmt.Fprintf(out, " ")
		}
		fmt.Fprintf(out, "%s", attr.Value)
	}
	return out.String()
}

// A little parser for Go FTL comment-directives.
type directive struct {
	Kind  string          `parser:"'ftl' ':' @Ident"`
	Attrs []directiveAttr `parser:"  @@*"`
}

func (f *ftlDirective) getPositional(i int) *directiveAttr {
	if i < len(f.attrs) && f.attrs[i].Key == nil {
		return &f.attrs[i]
	}
	return nil
}

func (f *ftlDirective) get(key string) *directiveAttr {
	for _, attr := range f.attrs {
		if attr.Key != nil && *attr.Key == key {
			return &attr
		}
	}
	return nil
}

type directiveAttr struct {
	Key   *string        `parser:"(@Ident '=')?"` // Key is optional.
	Value directiveValue `parser:"@@"`
}

type directiveValue struct {
	Bool  *dirBool `parser:"  @('true'|'false')"`
	Ident *string  `parser:"| @Ident"`
	Path  *string  `parser:"| @(('/' (('{' Ident '}') | Ident))+ '/'?)"`
	Str   *string  `parser:"| @String"`
	Int   *int64   `parser:"| @Int"`
	Float *float64 `parser:"| @Float"`
}

type dirBool bool

func (b *dirBool) UnmarshalText(d []byte) error {
	*b = dirBool(string(d) == "true")
	return nil
}

// Text returns the text representation of the directive value.
func (d directiveValue) Text() string {
	switch {
	case d.Ident != nil:
		return *d.Ident
	case d.Path != nil:
		return *d.Path
	case d.Str != nil:
		return *d.Str
	case d.Int != nil:
		return strconv.FormatInt(*d.Int, 10)
	case d.Float != nil:
		return strconv.FormatFloat(*d.Float, 'g', 2, 64)
	case d.Bool != nil:
		return strconv.FormatBool(bool(*d.Bool))
	default:
		panic("??")
	}
}

func (d directiveValue) String() string {
	switch {
	case d.Ident != nil:
		return *d.Ident
	case d.Path != nil:
		return *d.Path
	case d.Str != nil:
		return strconv.Quote(*d.Str)
	case d.Int != nil:
		return strconv.FormatInt(*d.Int, 10)
	case d.Float != nil:
		return strconv.FormatFloat(*d.Float, 'g', 2, 64)
	case d.Bool != nil:
		return strconv.FormatBool(bool(*d.Bool))
	default:
		panic("??")
	}
}

var directiveParser = participle.MustBuild[directive](participle.Unquote(), participle.UseLookahead(2))

func parseDirectives(fset *token.FileSet, docs *ast.CommentGroup) ([]ftlDirective, error) {
	if docs == nil {
		return nil, nil
	}
	var directives []ftlDirective
	for _, line := range docs.List {
		if !strings.HasPrefix(line.Text, "//ftl:") {
			continue
		}
		ast, err := directiveParser.ParseString("", line.Text[2:])
		if err != nil {
			// Adjust the Participle-reported position relative to the AST node.
			pos := fset.Position(line.Pos())
			var perr participle.Error
			if errors.As(err, &perr) {
				ppos := perr.Position()
				ppos.Filename = pos.Filename
				ppos.Column += pos.Column
				ppos.Line += pos.Line - 1
				err = participle.Errorf(ppos, "%s", perr.Message())
			} else {
				err = fmt.Errorf("%s: %w", pos, err)
			}
			return nil, fmt.Errorf("%s: %w", "invalid directive", err)
		}
		directives = append(directives, ftlDirective{
			kind:  ast.Kind,
			attrs: ast.Attrs,
			pos:   goPosToSchemaPos(line.Pos()),
		})
	}
	return directives, nil
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
