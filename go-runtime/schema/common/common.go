package common

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
)

var (
	// FtlUnitTypePath is the path to the FTL unit type.
	FtlUnitTypePath   = "github.com/TBD54566975/ftl/go-runtime/ftl.Unit"
	FtlOptionTypePath = "github.com/TBD54566975/ftl/go-runtime/ftl.Option"

	extractorRegistery = xsync.NewMapOf[reflect.Type, ExtractDeclFunc[schema.Decl, ast.Node]]()
)

func NewExtractor(name string, factType analysis.Fact, run func(*analysis.Pass) (interface{}, error)) *analysis.Analyzer {
	if !reflect.TypeOf(factType).Implements(reflect.TypeOf((*SchemaFact)(nil)).Elem()) {
		panic(fmt.Sprintf("factType %T does not implement SchemaFact", factType))
	}
	return &analysis.Analyzer{
		Name:             name,
		Doc:              fmt.Sprintf("extracts %s schema elements to the module", name),
		Run:              run,
		ResultType:       reflect.TypeFor[ExtractorResult](),
		RunDespiteErrors: true,
		FactTypes:        []analysis.Fact{factType},
	}
}

type ExtractDeclFunc[T schema.Decl, N ast.Node] func(pass *analysis.Pass, node N, object types.Object) optional.Option[T]

func NewDeclExtractor[T schema.Decl, N ast.Node](name string, extractFunc ExtractDeclFunc[T, N]) *analysis.Analyzer {
	type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
	dType := reflect.TypeFor[T]()
	if _, ok := extractorRegistery.Load(dType); ok {
		panic(fmt.Sprintf("multiple extractors registered for %s", dType.String()))
	}
	wrapped := func(pass *analysis.Pass, n ast.Node, o types.Object) optional.Option[schema.Decl] {
		decl, ok := extractFunc(pass, n.(N), o).Get()
		if ok {
			return optional.Some(schema.Decl(decl))
		}
		return optional.None[schema.Decl]()
	}
	extractorRegistery.Store(dType, wrapped)
	return NewExtractor(name, (*DefaultFact[Tag])(nil), runExtractDeclsFunc[T, N](extractFunc))
}

type ExtractorResult struct {
	Facts []analysis.ObjectFact
}

func NewExtractorResult(pass *analysis.Pass) ExtractorResult {
	return ExtractorResult{Facts: pass.AllObjectFacts()}
}

// runExtractDeclsFunc extracts schema declarations from the AST.
//
// The `extractFunc` function is called on each node and should return the schema declaration for that node.
// If the node does not represent a schema declaration, the function should return `optional.None[T]()`.
//
// Only nodes that have been marked with a `common.ExtractedMetadata` fact are considered for extraction (nodes
// explicitly annotated with an FTL directive). Implicit schema declarations are extracted by the `transitive`
// extractor.
func runExtractDeclsFunc[T schema.Decl, N ast.Node](extractFunc ExtractDeclFunc[T, N]) func(pass *analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		nodeFilter := []ast.Node{ //nolint:forcetypeassert
			reflect.New(reflect.TypeFor[N]().Elem()).Interface().(N),
		}
		in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
		in.Preorder(nodeFilter, func(n ast.Node) {
			obj, ok := GetObjectForNode(pass.TypesInfo, n).Get()
			if !ok {
				return
			}
			if obj != nil && !IsPathInPkg(pass.Pkg, obj.Pkg().Path()) {
				return
			}
			md, ok := GetFactForObject[*ExtractedMetadata](pass, obj).Get()
			if !ok {
				return
			}
			if _, ok = md.Type.(T); !ok {
				return
			}
			if decl, ok := extractFunc(pass, n.(N), obj).Get(); ok {
				MarkSchemaDecl(pass, obj, decl)
			} else {
				MarkFailedExtraction(pass, obj)
			}
		})
		return NewExtractorResult(pass), nil
	}
}

func ExtractComments(doc *ast.CommentGroup) []string {
	if doc == nil {
		return nil
	}
	comments := []string{}
	if doc := doc.Text(); doc != "" {
		comments = strings.Split(strings.TrimSpace(doc), "\n")
	}
	return comments
}

func ExtractType(pass *analysis.Pass, pos token.Pos, tnode types.Type) optional.Option[schema.Type] {
	if tnode == nil {
		return optional.None[schema.Type]()
	}

	fset := pass.Fset
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return optional.Some[schema.Type](&schema.Ref{Pos: GoPosToSchemaPos(fset, pos), Name: tparam.Obj().Id()})
	}

	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			return extractRef(pass, pos, named)
		}
		switch underlying.Kind() {
		case types.String:
			return optional.Some[schema.Type](&schema.String{Pos: GoPosToSchemaPos(fset, pos)})

		case types.Int, types.Int64:
			return optional.Some[schema.Type](&schema.Int{Pos: GoPosToSchemaPos(fset, pos)})

		case types.Bool:
			return optional.Some[schema.Type](&schema.Bool{Pos: GoPosToSchemaPos(fset, pos)})

		case types.Float64:
			return optional.Some[schema.Type](&schema.Float{Pos: GoPosToSchemaPos(fset, pos)})

		default:
			return optional.None[schema.Type]()
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			NoEndColumnErrorf(pass, pos, "expected named type but got %s", tnode)
			return optional.None[schema.Type]()
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return optional.Some[schema.Type](&schema.Time{Pos: GoPosToSchemaPos(fset, pos)})

		case FtlUnitTypePath:
			return optional.Some[schema.Type](&schema.Unit{Pos: GoPosToSchemaPos(fset, pos)})

		case FtlOptionTypePath:
			typ := ExtractType(pass, pos, named.TypeArgs().At(0))
			if underlying, ok := typ.Get(); ok {
				return optional.Some[schema.Type](&schema.Optional{Pos: GoPosToSchemaPos(pass.Fset, pos), Type: underlying})
			}
			return optional.None[schema.Type]()

		default:
			return extractRef(pass, pos, named)
		}

	case *types.Map:
		return extractMap(pass, pos, underlying)

	case *types.Slice:
		return extractSlice(pass, pos, underlying)

	case *types.Interface:
		if underlying.String() == "any" {
			return optional.Some[schema.Type](&schema.Any{Pos: GoPosToSchemaPos(fset, pos)})
		}
		if named, ok := tnode.(*types.Named); ok {
			return extractRef(pass, pos, named)
		}
		return optional.None[schema.Type]()

	default:
		return optional.None[schema.Type]()
	}
}

func ExtractFuncForDecl(t schema.Decl) (ExtractDeclFunc[schema.Decl, ast.Node], error) {
	if f, ok := extractorRegistery.Load(reflect.TypeOf(t)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("no extractor registered for %T", t)
}

func GoPosToSchemaPos(fset *token.FileSet, pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func FtlModuleFromGoPackage(pkgPath string) (string, error) {
	parts := strings.Split(pkgPath, "/")
	if parts[0] != "ftl" {
		return "", fmt.Errorf("package %q is not in the ftl namespace", pkgPath)
	}
	return strings.TrimSuffix(parts[1], "_test"), nil
}

func IsType[T types.Type](t types.Type) bool {
	if _, ok := t.(*types.Named); ok {
		t = t.Underlying()
	}
	_, ok := t.(T)
	return ok
}

func IsPathInPkg(pkg *types.Package, path string) bool {
	if path == pkg.Path() {
		return true
	}
	return strings.HasPrefix(path, pkg.Path()+"/")
}

func GetObjectForNode(typesInfo *types.Info, node ast.Node) optional.Option[types.Object] {
	var obj types.Object
	switch n := node.(type) {
	case *ast.GenDecl:
		if len(n.Specs) > 0 {
			return GetObjectForNode(typesInfo, n.Specs[0])
		}
	case *ast.Field:
		if len(n.Names) > 0 {
			obj = typesInfo.ObjectOf(n.Names[0])
		}
	case *ast.ImportSpec:
		obj = typesInfo.ObjectOf(n.Name)
	case *ast.ValueSpec:
		if len(n.Names) > 0 {
			obj = typesInfo.ObjectOf(n.Names[0])
		}
	case *ast.TypeSpec:
		obj = typesInfo.ObjectOf(n.Name)
	case *ast.FuncDecl:
		obj = typesInfo.ObjectOf(n.Name)
	default:
		return optional.None[types.Object]()
	}
	if obj == nil {
		return optional.None[types.Object]()
	}
	return optional.Some(obj)
}

func GetTypeForNode(node ast.Node, info *types.Info) types.Type {
	switch n := node.(type) {
	case *ast.Ident:
		if obj := info.ObjectOf(n); obj != nil {
			return obj.Type()
		}
	case *ast.AssignStmt:
		if len(n.Lhs) > 0 {
			return info.TypeOf(n.Lhs[0])
		}
	case *ast.ValueSpec:
		if len(n.Names) > 0 {
			if obj := info.ObjectOf(n.Names[0]); obj != nil {
				return obj.Type()
			}
		}
	case *ast.TypeSpec:
		return info.TypeOf(n.Type)
	case *ast.CompositeLit:
		return info.TypeOf(n)
	case *ast.CallExpr:
		return info.TypeOf(n)
	case *ast.FuncDecl:
		if n.Name != nil {
			if obj := info.ObjectOf(n.Name); obj != nil {
				return obj.Type()
			}
		}
	case *ast.GenDecl:
		for _, spec := range n.Specs {
			if t := GetTypeForNode(spec, info); t != nil {
				return t
			}
		}
	case ast.Expr:
		return info.TypeOf(n)
	}
	return nil
}

func extractRef(pass *analysis.Pass, pos token.Pos, named *types.Named) optional.Option[schema.Type] {
	if named.Obj().Pkg() == nil {
		return optional.None[schema.Type]()
	}

	nodePath := named.Obj().Pkg().Path()
	if !IsPathInPkg(pass.Pkg, nodePath) && !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
		NoEndColumnErrorf(pass, pos, "unsupported external type %q", named.Obj().Pkg().Path()+"."+named.Obj().Name())
		return optional.None[schema.Type]()
	}

	moduleName, err := FtlModuleFromGoPackage(nodePath)
	if err != nil {
		noEndColumnWrapf(pass, pos, err, "")
		return optional.None[schema.Type]()
	}

	ref := &schema.Ref{
		Pos:    GoPosToSchemaPos(pass.Fset, pos),
		Module: moduleName,
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
	}
	for i := range named.TypeArgs().Len() {
		typeArg, ok := ExtractType(pass, pos, named.TypeArgs().At(i)).Get()
		if !ok {
			TokenErrorf(pass, pos, named.TypeArgs().At(i).String(), "unsupported type %q for type argument", named.TypeArgs().At(i))
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

	if isLocalRef(pass, ref) {
		// mark this local reference to ensure its underlying schema type is hydrated by the appropriate extractor and
		// included in the schema
		markNeedsExtraction(pass, named.Obj())
	}

	return optional.Some[schema.Type](ref)
}

func extractMap(pass *analysis.Pass, pos token.Pos, tnode *types.Map) optional.Option[schema.Type] {
	key, ok := ExtractType(pass, pos, tnode.Key()).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	value, ok := ExtractType(pass, pos, tnode.Elem()).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Map{Pos: GoPosToSchemaPos(pass.Fset, pos), Key: key, Value: value})
}

func extractSlice(pass *analysis.Pass, pos token.Pos, tnode *types.Slice) optional.Option[schema.Type] {
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return optional.Some[schema.Type](&schema.Bytes{Pos: GoPosToSchemaPos(pass.Fset, pos)})
	}

	value, ok := ExtractType(pass, pos, tnode.Elem()).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Array{
		Pos:     GoPosToSchemaPos(pass.Fset, pos),
		Element: value,
	})
}

func ExtractTypeForNode(pass *analysis.Pass, obj types.Object, node ast.Node, index types.Type) optional.Option[schema.Type] {
	switch typ := node.(type) {
	// Selector expression e.g. ftl.Unit, ftl.Option, foo.Bar
	case *ast.SelectorExpr:
		var ident *ast.Ident
		var ok bool
		if ident, ok = typ.X.(*ast.Ident); !ok {
			return optional.None[schema.Type]()
		}

		for _, im := range pass.Pkg.Imports() {
			if im.Name() != ident.Name {
				continue
			}
			switch im.Path() /*"." + typ.Sel.Name */ {
			case "time.Time":
				return optional.Some[schema.Type](&schema.Time{})
			case FtlUnitTypePath:
				return optional.Some[schema.Type](&schema.Unit{})
			case FtlOptionTypePath:
				if index == nil {
					return optional.None[schema.Type]()
				}
				return ExtractType(pass, node.Pos(), index)
			default: // Data ref
				externalModuleName, err := FtlModuleFromGoPackage(im.Path())
				if err != nil {
					return optional.None[schema.Type]()
				}
				return optional.Some[schema.Type](&schema.Ref{
					Pos:    GoPosToSchemaPos(pass.Fset, node.Pos()),
					Module: externalModuleName,
					Name:   typ.Sel.Name,
				})
			}
		}

	case *ast.IndexExpr: // Generic type, e.g. ftl.Option[string]
		if se, ok := typ.X.(*ast.SelectorExpr); ok {
			return ExtractTypeForNode(pass, obj, se, pass.TypesInfo.TypeOf(typ.Index))
		}

	default:
		variantNode := GetTypeForNode(node, pass.TypesInfo)
		if _, ok := variantNode.(*types.Struct); ok {
			variantNode = obj.Type()
		}
		return ExtractType(pass, node.Pos(), variantNode)
	}

	return optional.None[schema.Type]()
}

func IsSelfReference(pass *analysis.Pass, obj types.Object, t schema.Type) bool {
	ref, ok := t.(*schema.Ref)
	if !ok {
		return false
	}
	moduleName, err := FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return false
	}
	return ref.Module == moduleName && strcase.ToUpperCamel(obj.Name()) == ref.Name
}

func isLocalRef(pass *analysis.Pass, ref *schema.Ref) bool {
	moduleName, err := FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return false
	}
	return ref.Module == "" || ref.Module == moduleName
}

func GetNativeName(obj types.Object) string {
	return obj.Pkg().Path() + "." + obj.Name()
}
