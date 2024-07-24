package common

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
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
	FtlUnitTypePath = "github.com/TBD54566975/ftl/go-runtime/ftl.Unit"
	// FtlOptionTypePath is the path to the FTL option type.
	FtlOptionTypePath = "github.com/TBD54566975/ftl/go-runtime/ftl.Option"

	extractorRegistery = xsync.NewMapOf[reflect.Type, ExtractDeclFunc[schema.Decl, ast.Node]]()
)

// NewExtractor creates a new schema element extractor.
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

// ExtractDeclFunc extracts a schema declaration from the given node.
type ExtractDeclFunc[T schema.Decl, N ast.Node] func(pass *analysis.Pass, node N, object types.Object) optional.Option[T]

// NewDeclExtractor creates a new schema declaration extractor and registers its extraction function with
// the common extractor registry.
// The registry provides functions for extracting schema declarations by type and is used to extract
// transitive declarations in a separate pass from the decl extraction pass.
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

// ExtractorResult contains the results of an extraction pass.
type ExtractorResult struct {
	Facts []analysis.ObjectFact
}

// NewExtractorResult creates a new ExtractorResult with all object facts from this pass.
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
			if obj != nil && !IsPathInModule(pass.Pkg, obj.Pkg().Path()) {
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

// ExtractComments extracts the comments from the given comment group.
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

// ExtractFuncForDecl returns the registered extraction function for the given declaration type.
func ExtractFuncForDecl(t schema.Decl) (ExtractDeclFunc[schema.Decl, ast.Node], error) {
	if f, ok := extractorRegistery.Load(reflect.TypeOf(t)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("no extractor registered for %T", t)
}

// GoPosToSchemaPos converts a Go token.Pos to a schema.Position.
func GoPosToSchemaPos(fset *token.FileSet, pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

// FtlModuleFromGoPackage returns the FTL module name from the given Go package path.
func FtlModuleFromGoPackage(pkgPath string) (string, error) {
	parts := strings.Split(pkgPath, "/")
	if parts[0] != "ftl" {
		return "", fmt.Errorf("package %q is not in the ftl namespace", pkgPath)
	}
	return strings.TrimSuffix(parts[1], "_test"), nil
}

// IsType returns true if the given type is of the specified type.
func IsType[T types.Type](t types.Type) bool {
	if _, ok := t.(*types.Named); ok {
		t = t.Underlying()
	}
	_, ok := t.(T)
	return ok
}

// IsPathInModule returns true if the given path is in the module.
func IsPathInModule(pkg *types.Package, path string) bool {
	if path == pkg.Path() {
		return true
	}
	moduleName, err := FtlModuleFromGoPackage(pkg.Path())
	if err != nil {
		return false
	}
	return strings.HasPrefix(path, "ftl/"+moduleName)
}

// ExtractType extracts the schema type for the given node.
func ExtractType(pass *analysis.Pass, node ast.Node) optional.Option[schema.Type] {
	tnode := GetTypeInfoForNode(node, pass.TypesInfo)
	externalType := extractExternalType(pass, node)
	if externalType.Ok() {
		return externalType
	}

	switch typ := node.(type) {
	case *ast.ArrayType:
		return extractSlice(pass, typ)

	case *ast.MapType:
		return extractMap(pass, typ)

	case *ast.InterfaceType:
		t, ok := tnode.Get()
		if !ok {
			return optional.None[schema.Type]()
		}
		iType, ok := t.Underlying().(*types.Interface)
		if !ok {
			return optional.None[schema.Type]()
		}
		if iType.Underlying().String() == "any" {
			return optional.Some[schema.Type](&schema.Any{Pos: GoPosToSchemaPos(pass.Fset, node.Pos())})
		}
		if _, ok := t.(*types.Named); ok {
			return extractRef(pass, node)
		}
		return optional.None[schema.Type]()

	case *ast.Field:
		return ExtractType(pass, typ.Type)

	case *ast.TypeSpec:
		if _, ok := typ.Type.(*ast.StructType); ok {
			return extractRef(pass, typ)
		}
		return ExtractType(pass, typ.Type)

	case *ast.Ident:
		if t, ok := tnode.Get(); ok {
			if tparam, ok := t.(*types.TypeParam); ok {
				return optional.Some[schema.Type](&schema.Ref{Pos: GoPosToSchemaPos(pass.Fset, node.Pos()), Name: tparam.Obj().Id()})
			}
			switch underlying := t.Underlying().(type) {
			case *types.Basic:
				if underlying.Kind() == types.Invalid {
					return optional.None[schema.Type]()
				}
				if _, ok := t.(*types.Named); ok {
					return extractRef(pass, node)
				}
				return extractBasicType(pass, node.Pos(), underlying)
			case *types.Interface:
				if underlying.String() == "any" {
					return optional.Some[schema.Type](&schema.Any{Pos: GoPosToSchemaPos(pass.Fset, node.Pos())})
				}
				if _, ok := t.(*types.Named); ok {
					return extractRef(pass, node)
				}
				return optional.None[schema.Type]()
			}

		}
		return extractRef(pass, typ)

	case *ast.SelectorExpr: // Selector expression e.g. ftl.Unit, ftl.Option, foo.Bar
		var ident *ast.Ident
		var ok bool
		if ident, ok = typ.X.(*ast.Ident); !ok {
			return optional.None[schema.Type]()
		}

		for _, im := range pass.Pkg.Imports() {
			if im.Name() != ident.Name {
				continue
			}

			switch im.Path() + "." + typ.Sel.Name {
			case "time.Time":
				return optional.Some[schema.Type](&schema.Time{})
			case FtlUnitTypePath:
				return optional.Some[schema.Type](&schema.Unit{})
			case FtlOptionTypePath:
				return optional.Some[schema.Type](&schema.Optional{
					Pos: GoPosToSchemaPos(pass.Fset, node.Pos()),
				})

			default: // Data ref
				if strings.HasPrefix(im.Path(), pass.Pkg.Path()+"/") {
					// subpackage, same module
					return ExtractType(pass, typ.Sel)
				}

				if !IsPathInModule(pass.Pkg, im.Path()) && IsExternalType(im.Path()) {
					NoEndColumnErrorf(pass, node.Pos(), "unsupported external type %q; see FTL docs on using external types: %s",
						im.Path()+"."+typ.Sel.Name, "tbd54566975.github.io/ftl/docs/reference/externaltypes/")
					return optional.None[schema.Type]()
				}

				// FTL, different module
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

	case *ast.IndexListExpr:
		t, ok := ExtractType(pass, typ.X).Get()
		if !ok {
			return optional.None[schema.Type]()
		}
		ref, ok := t.(*schema.Ref)
		if !ok {
			return optional.None[schema.Type]()
		}
		var params []schema.Type
		for _, idx := range typ.Indices {
			if param, ok := ExtractType(pass, idx).Get(); ok {
				params = append(params, param)
			}
		}
		ref.TypeParameters = params
		return optional.Some[schema.Type](ref)

	case *ast.IndexExpr: // Generic type, e.g. ftl.Option[string]
		t, ok := ExtractType(pass, typ.X).Get()
		if !ok {
			return optional.None[schema.Type]()
		}
		idx, ok := ExtractType(pass, typ.Index).Get()
		if !ok {
			return optional.None[schema.Type]()
		}
		switch s := t.(type) {
		case *schema.Ref:
			s.TypeParameters = []schema.Type{idx}
		case *schema.Optional:
			s.Type = idx
		default:
			return optional.None[schema.Type]()
		}
		return optional.Some[schema.Type](t)
	}

	return optional.None[schema.Type]()
}

// extracts a ref to the type alias over an external type
func extractExternalType(pass *analysis.Pass, node ast.Node) optional.Option[schema.Type] {
	obj, ok := GetObjectForNode(pass.TypesInfo, node).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	tn, ok := obj.(*types.TypeName)
	if !ok {
		return optional.None[schema.Type]()
	}

	if tn.Pkg() == nil {
		return optional.None[schema.Type]()
	}

	moduleName, err := FtlModuleFromGoPackage(tn.Pkg().Path())
	if err != nil {
		return optional.None[schema.Type]()
	}
	currentModule, err := FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return optional.None[schema.Type]()
	}
	if underlying, ok := obj.Type().(*types.Named); ok &&
		moduleName == currentModule && // type is in this module
		IsExternalType(underlying.Obj().Pkg().Path()) { // alias— e.g. type MyType = foo.OtherType
		MarkNeedsExtraction(pass, obj)
		return optional.Some[schema.Type](&schema.Ref{
			Pos:    GoPosToSchemaPos(pass.Fset, node.Pos()),
			Module: moduleName,
			Name:   strcase.ToUpperCamel(obj.Name()),
		})
	}
	return optional.None[schema.Type]()
}

func extractBasicType(pass *analysis.Pass, pos token.Pos, basic *types.Basic) optional.Option[schema.Type] {
	switch basic.Kind() {
	case types.String:
		return optional.Some[schema.Type](&schema.String{Pos: GoPosToSchemaPos(pass.Fset, pos)})

	case types.Int:
		return optional.Some[schema.Type](&schema.Int{Pos: GoPosToSchemaPos(pass.Fset, pos)})

	case types.Bool:
		return optional.Some[schema.Type](&schema.Bool{Pos: GoPosToSchemaPos(pass.Fset, pos)})

	case types.Float64:
		return optional.Some[schema.Type](&schema.Float{Pos: GoPosToSchemaPos(pass.Fset, pos)})

	default:
		return optional.None[schema.Type]()
	}
}

func extractRef(pass *analysis.Pass, node ast.Node) optional.Option[schema.Type] {
	obj, ok := GetObjectForNode(pass.TypesInfo, node).Get()
	if !ok {
		return optional.None[schema.Type]()
	}
	if obj.Pkg() == nil {
		return optional.None[schema.Type]()
	}

	nodePath := obj.Pkg().Path()
	if !IsPathInModule(pass.Pkg, nodePath) && IsExternalType(nodePath) {
		NoEndColumnErrorf(pass, node.Pos(), "unsupported external type %q; see FTL docs on using external types: %s",
			GetNativeName(obj), "tbd54566975.github.io/ftl/docs/reference/externaltypes/")
		return optional.None[schema.Type]()
	}

	moduleName, err := FtlModuleFromGoPackage(nodePath)
	if err != nil {
		noEndColumnWrapf(pass, node.Pos(), err, "")
		return optional.None[schema.Type]()
	}

	ref := &schema.Ref{
		Pos:    GoPosToSchemaPos(pass.Fset, node.Pos()),
		Module: moduleName,
	}
	if t, ok := node.(*ast.TypeSpec); ok {
		if t.TypeParams != nil {
			for _, p := range t.TypeParams.List {
				param, ok := ExtractType(pass, p).Get()
				var typename string
				if t, ok := GetTypeInfoForNode(p, pass.TypesInfo).Get(); ok {
					typename = fmt.Sprintf("%q ", t.String())
				}
				if !ok {
					Errorf(pass, p, "unsupported type %sfor type argument", typename)
					continue
				}

				// Fully qualify the Ref if needed
				if r, okArg := param.(*schema.Ref); okArg {
					if r.Module == "" {
						r.Module = moduleName
					}
					param = r
				}
				ref.TypeParameters = append(ref.TypeParameters, param)
			}
		}
	}
	ref.Name = strcase.ToUpperCamel(getNodeName(node))
	if ref.Name == "" {
		return optional.None[schema.Type]()
	}

	if isLocalRef(pass, ref) {
		// mark this local reference to ensure its underlying schema type is hydrated by the appropriate extractor and
		// included in the schema
		MarkNeedsExtraction(pass, obj)
	}

	return optional.Some[schema.Type](ref)
}

func extractMap(pass *analysis.Pass, node *ast.MapType) optional.Option[schema.Type] {
	key, ok := ExtractType(pass, node.Key).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	value, ok := ExtractType(pass, node.Value).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Map{Pos: GoPosToSchemaPos(pass.Fset, node.Pos()), Key: key, Value: value})
}

func extractSlice(pass *analysis.Pass, node *ast.ArrayType) optional.Option[schema.Type] {
	typ, ok := GetTypeInfoForNode(node, pass.TypesInfo).Get()
	if !ok {
		return optional.None[schema.Type]()
	}
	tnode, ok := typ.(*types.Slice)
	if !ok {
		return optional.None[schema.Type]()
	}
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return optional.Some[schema.Type](&schema.Bytes{Pos: GoPosToSchemaPos(pass.Fset, node.Pos())})
	}

	value, ok := ExtractType(pass, node.Elt).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Array{
		Pos:     GoPosToSchemaPos(pass.Fset, node.Pos()),
		Element: value,
	})
}

func getNodeName(node ast.Node) string {
	switch t := node.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.TypeSpec:
		return t.Name.Name
	case *ast.Field:
		if len(t.Names) > 0 {
			return getNodeName(t.Names[0])
		}
	}
	return ""
}

// GetObjectForNode returns the types.Object for the given node.
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
	case *ast.Ident:
		obj = typesInfo.ObjectOf(n)
	case *ast.IndexExpr:
		return GetObjectForNode(typesInfo, n.X)
	case *ast.IndexListExpr:
		return GetObjectForNode(typesInfo, n.X)
	case *ast.SelectorExpr:
		return GetObjectForNode(typesInfo, n.Sel)
	case *ast.ArrayType:
		return GetObjectForNode(typesInfo, n.Elt)
	default:
		return optional.None[types.Object]()
	}
	if obj == nil {
		return optional.None[types.Object]()
	}
	return optional.Some(obj)
}

func GetTypeInfoForNode(node ast.Node, info *types.Info) optional.Option[types.Type] {
	switch n := node.(type) {
	case *ast.Ident:
		if obj := info.ObjectOf(n); obj != nil {
			return optional.Some(obj.Type())
		}
	case *ast.AssignStmt:
		if len(n.Lhs) > 0 {
			return optional.Some(info.TypeOf(n.Lhs[0]))
		}
	case *ast.ValueSpec:
		if len(n.Names) > 0 {
			if obj := info.ObjectOf(n.Names[0]); obj != nil {
				return optional.Some(obj.Type())
			}
		}
	case *ast.TypeSpec:
		return optional.Some(info.TypeOf(n.Type))
	case *ast.CompositeLit:
		return optional.Some(info.TypeOf(n))
	case *ast.CallExpr:
		return optional.Some(info.TypeOf(n))
	case *ast.FuncDecl:
		if n.Name != nil {
			if obj := info.ObjectOf(n.Name); obj != nil {
				return optional.Some(obj.Type())
			}
		}
	case *ast.GenDecl:
		for _, spec := range n.Specs {
			if t := GetTypeInfoForNode(spec, info); t.Ok() {
				return t
			}
		}
	case *ast.Field:
		return optional.Some(info.TypeOf(n.Type))
	case *ast.SliceExpr:
		return optional.Some(info.TypeOf(n))
	case ast.Expr:
		return optional.Some(info.TypeOf(n))

	}
	return optional.None[types.Type]()
}

// IsSelfReference returns true if the schema reference refers to this object itself.
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

// GetNativeName returns the fully qualified name of the object, e.g. "github.com/TBD54566975/ftl/go-runtime/ftl.Unit".
func GetNativeName(obj types.Object) string {
	fqName := obj.Pkg().Path()
	if parts := strings.Split(obj.Pkg().Path(), "/"); parts[len(parts)-1] != obj.Pkg().Name() {
		fqName = fqName + "." + obj.Pkg().Name()
	}
	return fqName + "." + obj.Name()
}

// IsExternalType returns true if the object is from an external package.
func IsExternalType(path string) bool {
	return !strings.HasPrefix(path, "ftl/") &&
		path != "time.Time" &&
		path != FtlUnitTypePath &&
		path != FtlOptionTypePath
}

// GetDeclTypeName returns the name of the declaration type, e.g. "verb" for *schema.Verb.
func GetDeclTypeName(d schema.Decl) string {
	typeStr := reflect.TypeOf(d).String()
	lastDotIndex := strings.LastIndex(typeStr, ".")
	if lastDotIndex == -1 {
		return typeStr
	}
	return strcase.ToLowerCamel(typeStr[lastDotIndex+1:])
}

func Deref[T types.Object](pass *analysis.Pass, node ast.Expr) (string, T) {
	var obj T
	switch node := node.(type) {
	case *ast.Ident:
		obj, _ = pass.TypesInfo.Uses[node].(T)
		return "", obj

	case *ast.SelectorExpr:
		x, ok := node.X.(*ast.Ident)
		if !ok {
			return "", obj
		}
		obj, _ = pass.TypesInfo.Uses[node.Sel].(T)
		return x.Name, obj

	case *ast.IndexExpr:
		return Deref[T](pass, node.X)

	default:
		return "", obj
	}
}

// CallExprFromVar extracts a call expression from a variable declaration, if present.
func CallExprFromVar(node *ast.GenDecl) optional.Option[*ast.CallExpr] {
	if node.Tok != token.VAR {
		return optional.None[*ast.CallExpr]()
	}
	if len(node.Specs) != 1 {
		return optional.None[*ast.CallExpr]()
	}
	vs, ok := node.Specs[0].(*ast.ValueSpec)
	if !ok {
		return optional.None[*ast.CallExpr]()
	}
	if len(vs.Values) != 1 {
		return optional.None[*ast.CallExpr]()
	}
	callExpr, ok := vs.Values[0].(*ast.CallExpr)
	if !ok {
		return optional.None[*ast.CallExpr]()
	}
	return optional.Some(callExpr)
}

// FuncPathEquals checks if the function call expression is a call to the given path.
func FuncPathEquals(pass *analysis.Pass, callExpr *ast.CallExpr, path string) bool {
	_, fn := Deref[*types.Func](pass, callExpr.Fun)
	if fn == nil {
		return false
	}
	if fn.FullName() != path {
		return false
	}
	return fn.FullName() == path
}

// ApplyMetadata applies the extracted metadata to the object, if present. Returns true if metadata was found and
// applied.
func ApplyMetadata[T schema.Decl](pass *analysis.Pass, obj types.Object, apply func(md *ExtractedMetadata)) bool {
	if md, ok := GetFactForObject[*ExtractedMetadata](pass, obj).Get(); ok {
		if _, ok = md.Type.(T); !ok && md.Type != nil {
			return false
		}
		apply(md)
		return true
	}
	return false
}

// ExtractStringLiteralArg extracts a string literal argument from a call expression at the given index.
func ExtractStringLiteralArg(pass *analysis.Pass, node *ast.CallExpr, argIndex int) string {
	if argIndex >= len(node.Args) {
		Errorf(pass, node, "expected string argument at index %d", argIndex)
		return ""
	}

	literal, ok := node.Args[argIndex].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		Errorf(pass, node, "expected string literal for argument at index %d", argIndex)
		return ""
	}

	s, err := strconv.Unquote(literal.Value)
	if err != nil {
		Wrapf(pass, node, err, "")
		return ""
	}
	if s == "" {
		Errorf(pass, node, "expected non-empty string literal for argument at index %d", argIndex)
		return ""
	}
	return s
}

func isLocalRef(pass *analysis.Pass, ref *schema.Ref) bool {
	moduleName, err := FtlModuleFromGoPackage(pass.Pkg.Path())
	if err != nil {
		return false
	}
	return ref.Module == "" || ref.Module == moduleName
}
