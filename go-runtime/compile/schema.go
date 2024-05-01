package compile

import (
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
	"github.com/TBD54566975/ftl/internal/errors"
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

	ftlCallFuncPath       = "github.com/TBD54566975/ftl/go-runtime/ftl.Call"
	ftlConfigFuncPath     = "github.com/TBD54566975/ftl/go-runtime/ftl.Config"
	ftlSecretFuncPath     = "github.com/TBD54566975/ftl/go-runtime/ftl.Secret" //nolint:gosec
	ftlPostgresDBFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.PostgresDatabase"
	ftlUnitTypePath       = "github.com/TBD54566975/ftl/go-runtime/ftl.Unit"
	aliasFieldTag         = "json"
)

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Decl]string

type enums map[string]*schema.Enum
type enumInterfaces map[string]*types.Interface

func noEndColumnErrorf(pos token.Pos, format string, args ...interface{}) *schema.Error {
	return tokenErrorf(pos, "", format, args...)
}

func tokenErrorf(pos token.Pos, tokenText string, format string, args ...interface{}) *schema.Error {
	goPos := goPosToSchemaPos(pos)
	endColumn := goPos.Column
	if len(tokenText) > 0 {
		endColumn += utf8.RuneCountInString(tokenText)
	}
	return schema.Errorf(goPosToSchemaPos(pos), endColumn, format, args...)
}

func errorf(node ast.Node, format string, args ...interface{}) *schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Errorf(pos, endCol, format, args...)
}

func tokenWrapf(pos token.Pos, tokenText string, err error, format string, args ...interface{}) *schema.Error {
	goPos := goPosToSchemaPos(pos)
	endColumn := goPos.Column
	if len(tokenText) > 0 {
		endColumn += utf8.RuneCountInString(tokenText)
	}
	return schema.Wrapf(goPos, endColumn, err, format, args...)
}

func wrapf(node ast.Node, err error, format string, args ...interface{}) *schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Wrapf(pos, endCol, err, format, args...)
}

type errorSet map[string]*schema.Error

func (e errorSet) add(err *schema.Error) {
	e[err.Error()] = err
}

func (e errorSet) addAll(errs ...*schema.Error) {
	for _, err := range errs {
		e.add(err)
	}
}

// ExtractModuleSchema statically parses Go FTL module source into a schema.Module.
func ExtractModuleSchema(dir string) (NativeNames, *schema.Module, []*schema.Error /*schema errors*/, error /*exceptions*/) {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dir,
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, "./...")
	if err != nil {
		return nil, nil, nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil, nil, fmt.Errorf("no packages found in %q, does \"go mod tidy\" need to be run?", dir)
	}
	nativeNames := NativeNames{}
	// Find module name
	module := &schema.Module{}
	merr := []error{}
	schemaErrs := []*schema.Error{}
	for _, pkg := range pkgs {
		moduleName, ok := ftlModuleFromGoModule(pkg.PkgPath).Get()
		if !ok {
			return nil, nil, nil, fmt.Errorf("package %q is not in the ftl namespace", pkg.PkgPath)
		}
		module.Name = moduleName
		if len(pkg.Errors) > 0 {
			for _, perr := range pkg.Errors {
				merr = append(merr, perr)
			}
		}
		pctx := &parseContext{pkg: pkg, pkgs: pkgs, module: module, nativeNames: NativeNames{}, enums: enums{}, enumInterfaces: enumInterfaces{}, errors: errorSet{}}
		for _, file := range pkg.Syntax {
			err := goast.Visit(file, func(node ast.Node, next func() error) (err error) {
				switch node := node.(type) {
				case *ast.CallExpr:
					visitCallExpr(pctx, node)

				case *ast.File:
					visitFile(pctx, node)

				case *ast.FuncDecl:
					verb := visitFuncDecl(pctx, node)
					pctx.activeVerb = verb
					err = next()
					if err != nil {
						return err
					}
					pctx.activeVerb = nil
					return nil

				case *ast.GenDecl:
					visitGenDecl(pctx, node)

				default:
				}
				return next()
			})
			if err != nil {
				return nil, nil, nil, err
			}
		}
		for decl, nativeName := range pctx.nativeNames {
			nativeNames[decl] = nativeName
		}
		for _, e := range maps.Values(pctx.enums) {
			pctx.module.Decls = append(pctx.module.Decls, e)
		}
		if len(pctx.errors) > 0 {
			schemaErrs = append(schemaErrs, maps.Values(pctx.errors)...)
		}
	}
	if len(schemaErrs) > 0 {
		schema.SortErrorsByPosition(schemaErrs)
		return nil, nil, schemaErrs, nil
	}
	if len(merr) > 0 {
		return nil, nil, nil, errors.Join(merr...)
	}
	return nativeNames, module, nil, schema.ValidateModule(module)
}

func visitCallExpr(pctx *parseContext, node *ast.CallExpr) {
	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return
	}
	switch fn.FullName() {
	case ftlCallFuncPath:
		parseCall(pctx, node)

	case ftlConfigFuncPath, ftlSecretFuncPath:
		// Secret/config declaration: ftl.Config[<type>](<name>)
		parseConfigDecl(pctx, node, fn)
	case ftlPostgresDBFuncPath:
		parseDatabaseDecl(pctx, node, schema.PostgresDatabaseType)
	}
}

func parseCall(pctx *parseContext, node *ast.CallExpr) {
	if len(node.Args) != 3 {
		pctx.errors.add(errorf(node, "call must have exactly three arguments"))
		return
	}
	_, verbFn := deref[*types.Func](pctx.pkg, node.Args[1])
	if verbFn == nil {
		if sel, ok := node.Args[1].(*ast.SelectorExpr); ok {
			pctx.errors.add(errorf(node.Args[1], "call first argument must be a function but is an unresolved reference to %s.%s", sel.X, sel.Sel))
		}
		pctx.errors.add(errorf(node.Args[1], "call first argument must be a function"))
		return
	}
	if pctx.activeVerb == nil {
		return
	}
	moduleName, ok := ftlModuleFromGoModule(verbFn.Pkg().Path()).Get()
	if !ok {
		pctx.errors.add(errorf(node.Args[1], "call first argument must be a function in an ftl module"))
		return
	}
	ref := &schema.Ref{
		Pos:    goPosToSchemaPos(node.Pos()),
		Module: moduleName,
		Name:   strcase.ToLowerCamel(verbFn.Name()),
	}
	pctx.activeVerb.AddCall(ref)
}

func parseConfigDecl(pctx *parseContext, node *ast.CallExpr, fn *types.Func) {
	var name string
	if len(node.Args) == 1 {
		if literal, ok := node.Args[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
			var err error
			name, err = strconv.Unquote(literal.Value)
			if err != nil {
				pctx.errors.add(wrapf(node, err, ""))
				return
			}
		}
	}
	if name == "" {
		pctx.errors.add(errorf(node, "config and secret declarations must have a single string literal argument"))
		return
	}
	index := node.Fun.(*ast.IndexExpr) //nolint:forcetypeassert

	// Type parameter
	tp := pctx.pkg.TypesInfo.Types[index.Index].Type
	st, ok := visitType(pctx, index.Index.Pos(), tp, false).Get()
	if !ok {
		pctx.errors.add(errorf(index.Index, "unsupported type %q", tp))
		return
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

	// Check for duplicates
	_, endCol := goNodePosToSchemaPos(node)
	for _, d := range pctx.module.Decls {
		switch fn.FullName() {
		case ftlConfigFuncPath:
			c, ok := d.(*schema.Config)
			if ok && c.Name == name && c.Type.String() == st.String() {
				pctx.errors.add(errorf(node, "duplicate config declaration at %d:%d-%d", c.Pos.Line, c.Pos.Column, endCol))
				return
			}
		case ftlSecretFuncPath:
			s, ok := d.(*schema.Secret)
			if ok && s.Name == name && s.Type.String() == st.String() {
				pctx.errors.add(errorf(node, "duplicate secret declaration at %d:%d-%d", s.Pos.Line, s.Pos.Column, endCol))
				return
			}
		default:
		}
	}

	pctx.module.Decls = append(pctx.module.Decls, decl)
}

func parseDatabaseDecl(pctx *parseContext, node *ast.CallExpr, dbType string) {
	var name string
	if len(node.Args) == 1 {
		if literal, ok := node.Args[0].(*ast.BasicLit); ok && literal.Kind == token.STRING {
			var err error
			name, err = strconv.Unquote(literal.Value)
			if err != nil {
				pctx.errors.add(wrapf(node, err, ""))
				return
			}
		}
	}
	if name == "" {
		pctx.errors.add(errorf(node, "config and secret declarations must have a single string literal argument"))
		return
	}

	// Check for duplicates
	_, endCol := goNodePosToSchemaPos(node)
	for _, d := range pctx.module.Decls {
		db, ok := d.(*schema.Database)
		if ok && db.Name == name {
			pctx.errors.add(errorf(node, "duplicate database declaration at %d:%d-%d", db.Pos.Line, db.Pos.Column, endCol))
			return
		}
	}

	decl := &schema.Database{
		Pos:  goPosToSchemaPos(node.Pos()),
		Name: name,
		Type: dbType,
	}
	pctx.module.Decls = append(pctx.module.Decls, decl)
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

func checkSignature(pctx *parseContext, node *ast.FuncDecl, sig *types.Signature) (req, resp optional.Option[*types.Var]) {
	params := sig.Params()
	results := sig.Results()

	if params.Len() > 2 {
		pctx.errors.add(errorf(node, "must have at most two parameters (context.Context, struct)"))
	}
	if params.Len() == 0 {
		pctx.errors.add(errorf(node, "first parameter must be context.Context"))
	} else if !types.AssertableTo(contextIfaceType(), params.At(0).Type()) {
		pctx.errors.add(tokenErrorf(params.At(0).Pos(), params.At(0).Name(), "first parameter must be of type context.Context but is %s", params.At(0).Type()))
	}

	if params.Len() == 2 {
		if !isType[*types.Struct](params.At(1).Type()) {
			pctx.errors.add(tokenErrorf(params.At(1).Pos(), params.At(1).Name(), "second parameter must be a struct but is %s", params.At(1).Type()))
		}
		if params.At(1).Type().String() == ftlUnitTypePath {
			pctx.errors.add(tokenErrorf(params.At(1).Pos(), params.At(1).Name(), "second parameter must not be ftl.Unit"))
		}

		req = optional.Some(params.At(1))
	}

	if results.Len() > 2 {
		pctx.errors.add(errorf(node, "must have at most two results (struct, error)"))
	}
	if results.Len() == 0 {
		pctx.errors.add(errorf(node, "must at least return an error"))
	} else if !types.AssertableTo(errorIFaceType(), results.At(results.Len()-1).Type()) {
		pctx.errors.add(tokenErrorf(results.At(results.Len()-1).Pos(), results.At(results.Len()-1).Name(), "must return an error but is %s", results.At(0).Type()))
	}
	if results.Len() == 2 {
		if !isType[*types.Struct](results.At(0).Type()) {
			pctx.errors.add(tokenErrorf(results.At(0).Pos(), results.At(0).Name(), "first result must be a struct but is %s", results.At(0).Type()))
		}
		if results.At(1).Type().String() == ftlUnitTypePath {
			pctx.errors.add(tokenErrorf(results.At(1).Pos(), results.At(1).Name(), "second result must not be ftl.Unit"))
		}
		resp = optional.Some(results.At(0))
	}
	return req, resp
}

func goPosToSchemaPos(pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func goNodePosToSchemaPos(node ast.Node) (schema.Position, int) {
	p := fset.Position(node.Pos())
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}, fset.Position(node.End()).Column
}

func maybeVisitTypeEnumVariant(pctx *parseContext, node *ast.GenDecl, isExported bool) bool {
	if len(node.Specs) != 1 {
		return false
	}
	// `type NAME TYPE` e.g. type Scalar string
	if t, ok := node.Specs[0].(*ast.TypeSpec); ok {
		enumVariant := &schema.EnumVariant{
			Pos:      goPosToSchemaPos(node.Pos()),
			Comments: visitComments(node.Doc),
			Name:     strcase.ToUpperCamel(t.Name.Name),
		}
		for enumName, interfaceNode := range pctx.enumInterfaces {
			// If the type declared is an enum variant, then it must implement
			// the interface of a type enum we've already read into pctx.enums
			// and pctx.enumInterfaces.
			if named, ok := pctx.pkg.Types.Scope().Lookup(t.Name.Name).Type().(*types.Named); ok {
				if types.Implements(named, interfaceNode) {
					if typ, ok := visitType(pctx, node.Pos(), named, isExported).Get(); ok {
						enumVariant.Value = &schema.TypeValue{Value: typ}
					} else {
						pctx.errors.add(errorf(node, "unsupported type %q for type enum variant", named))
					}
					pctx.enums[enumName].Variants = append(pctx.enums[enumName].Variants, enumVariant)
					return true
				}
			}
		}
	}
	return false
}

func visitGenDecl(pctx *parseContext, node *ast.GenDecl) {
	switch node.Tok {
	case token.TYPE:
		if node.Doc == nil {
			_ = maybeVisitTypeEnumVariant(pctx, node, false)
			return
		}
		directives, err := parseDirectives(node, fset, node.Doc)
		if err != nil {
			pctx.errors.add(err)
		}

		// If any directives on this node are exported, then the node is
		// considered exported for type enum variant purposes
		enumVarIsExported := false
		for _, dir := range directives {
			if exportableDir, ok := dir.(exportable); ok {
				enumVarIsExported = enumVarIsExported || exportableDir.IsExported()
			}
		}
		if maybeVisitTypeEnumVariant(pctx, node, enumVarIsExported) {
			return
		}

		for _, dir := range directives {
			switch dir.(type) {
			case *directiveVerb, *directiveData, *directiveEnum:
				if len(node.Specs) != 1 {
					pctx.errors.add(errorf(node, "error parsing ftl directive: expected "+
						"exactly one type declaration"))
					return
				}
				if pctx.module.Name == "" {
					pctx.module.Name = pctx.pkg.Name
				} else if pctx.module.Name != pctx.pkg.Name && strings.TrimPrefix(pctx.pkg.Name, "ftl/") != pctx.module.Name {
					pctx.errors.add(errorf(node, "ftl directive must be in the module package"))
					return
				}
				isExported := false
				if exportableDir, ok := dir.(exportable); ok {
					isExported = exportableDir.IsExported()
				}
				if t, ok := node.Specs[0].(*ast.TypeSpec); ok {
					typ := pctx.pkg.TypesInfo.TypeOf(t.Type)
					switch typ.Underlying().(type) {
					case *types.Basic:
						if typ, ok := visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.TypeOf(t.Type), isExported).Get(); ok {
							enum := &schema.Enum{
								Pos:      goPosToSchemaPos(node.Pos()),
								Comments: visitComments(node.Doc),
								Name:     strcase.ToUpperCamel(t.Name.Name),
								Type:     typ,
								Export:   isExported,
							}
							pctx.enums[t.Name.Name] = enum
							pctx.nativeNames[enum] = t.Name.Name
						} else {
							pctx.errors.add(errorf(node, "unsupported type %q for value enum",
								pctx.pkg.TypesInfo.TypeOf(t.Type).Underlying()))
						}
					case *types.Interface:
						enum := &schema.Enum{
							Pos:      goPosToSchemaPos(node.Pos()),
							Comments: visitComments(node.Doc),
							Name:     strcase.ToUpperCamel(t.Name.Name),
							Export:   isExported,
						}
						pctx.enums[t.Name.Name] = enum
						pctx.nativeNames[enum] = t.Name.Name
						if typ, ok := typ.(*types.Interface); ok {
							pctx.enumInterfaces[t.Name.Name] = typ
						} else {
							pctx.errors.add(errorf(node, "expected interface for type enum but got %q", typ))
						}
					}
					visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.Defs[t.Name].Type(), isExported)
				}

			case *directiveIngress, *directiveCronJob:
			}
		}
		return

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
			visitValueSpec(pctx, v)
		}
		return

	default:
		return
	}
}

// maybeErrorOnInvalidEnumMixing ensures value enums are not set as variants of type enums.
// How this gets parsed:
//
// //ftl:enum
// type TypeEnum interface { typeEnum() }
//
// type BadValueEnum int
//
// // This line causes BadValueEnum to be parsed as a TypeEnum variant. At this point, we
// // cannot determine if BadValueEnum is intended to be a value enum, so we must treat it
// // as any other type enum variant.
// func (BadValueEnum) typeEnum() {}
//
// // This line will error because this is where we find out that BadValueEnum is intended
// // to be a value enum, but value enums cannot be variants of type enums. BadValueEnum
// // is not in pctx.enums.
// const A BadValueEnum = 1
func maybeErrorOnInvalidEnumMixing(pctx *parseContext, node *ast.ValueSpec, enumName string) {
	for _, enum := range pctx.enums {
		for _, variant := range enum.Variants {
			if variant.Name == enumName {
				pctx.errors.add(errorf(node, "cannot attach enum value to %s because it is a variant of type enum %s, not a value enum", enumName, enum.Name))
			}
		}
	}
}

func visitValueSpec(pctx *parseContext, node *ast.ValueSpec) {
	var enum *schema.Enum
	i, ok := node.Type.(*ast.Ident)
	if ok {
		enum = pctx.enums[i.Name]
	}
	if enum == nil {
		maybeErrorOnInvalidEnumMixing(pctx, node, i.Name)
		return
	}
	c, ok := pctx.pkg.TypesInfo.Defs[node.Names[0]].(*types.Const)
	if !ok {
		pctx.errors.add(errorf(node, "could not extract enum %s: expected exactly one variant name", enum.Name))
		return
	}

	if value, ok := visitConst(pctx, c).Get(); ok {
		variant := &schema.EnumVariant{
			Pos:      goPosToSchemaPos(c.Pos()),
			Comments: visitComments(node.Doc),
			Name:     strcase.ToUpperCamel(c.Id()),
			Value:    value,
		}
		enum.Variants = append(enum.Variants, variant)
	} else {
		pctx.errors.add(errorf(node, "unsupported type %q for enum variant %q", c.Type(), c.Name()))
	}
}

func visitFuncDecl(pctx *parseContext, node *ast.FuncDecl) (verb *schema.Verb) {
	if node.Doc == nil {
		return nil
	}
	directives, err := parseDirectives(node, fset, node.Doc)
	if err != nil {
		pctx.errors.add(err)
	}
	var metadata []schema.Metadata
	isVerb := false
	isExported := false
	for _, dir := range directives {
		switch dir := dir.(type) {
		case *directiveVerb:
			isVerb = true
			isExported = dir.Export
			if pctx.module.Name == "" {
				pctx.module.Name = pctx.pkg.Name
			} else if pctx.module.Name != pctx.pkg.Name {
				pctx.errors.add(errorf(node, "function verb directive must be in the module package"))
			}
		case *directiveIngress:
			isVerb = true
			isExported = true
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
		case *directiveCronJob:
			isVerb = true
			isExported = false
			metadata = append(metadata, &schema.MetadataCronJob{
				Pos:  dir.Pos,
				Cron: dir.Cron,
			})
		case *directiveData, *directiveEnum:
			pctx.errors.add(errorf(node, "unexpected directive %T", dir))
		}
	}
	if !isVerb {
		return nil
	}

	for _, name := range pctx.nativeNames {
		if name == node.Name.Name {
			pctx.errors.add(noEndColumnErrorf(node.Pos(), "duplicate verb name %q", node.Name.Name))
			return nil
		}
	}

	fnt := pctx.pkg.TypesInfo.Defs[node.Name].(*types.Func) //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature)                    //nolint:forcetypeassert
	if sig.Recv() != nil {
		pctx.errors.add(errorf(node, "ftl:verb cannot be a method"))
		return nil
	}
	params := sig.Params()
	results := sig.Results()
	reqt, respt := checkSignature(pctx, node, sig)

	var req optional.Option[schema.Type]
	if reqt.Ok() {
		req = visitType(pctx, node.Pos(), params.At(1).Type(), isExported)
	} else {
		req = optional.Some[schema.Type](&schema.Unit{})
	}
	var resp optional.Option[schema.Type]
	if respt.Ok() {
		resp = visitType(pctx, node.Pos(), results.At(0).Type(), isExported)
	} else {
		resp = optional.Some[schema.Type](&schema.Unit{})
	}
	reqV, reqOk := req.Get()
	resV, respOk := resp.Get()
	if !reqOk {
		pctx.errors.add(tokenErrorf(params.At(1).Pos(), params.At(1).Name(),
			"unsupported request type %q", params.At(1).Type()))
	}
	if !respOk {
		pctx.errors.add(tokenErrorf(results.At(0).Pos(), results.At(0).Name(),
			"unsupported response type %q", results.At(0).Type()))
	}
	verb = &schema.Verb{
		Pos:      goPosToSchemaPos(node.Pos()),
		Comments: visitComments(node.Doc),
		Export:   isExported,
		Name:     strcase.ToLowerCamel(node.Name.Name),
		Request:  reqV,
		Response: resV,
		Metadata: metadata,
	}
	pctx.nativeNames[verb] = node.Name.Name
	pctx.module.Decls = append(pctx.module.Decls, verb)
	return verb
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

func visitStruct(pctx *parseContext, pos token.Pos, tnode types.Type, isExported bool) optional.Option[*schema.Ref] {
	named, ok := tnode.(*types.Named)
	if !ok {
		pctx.errors.add(noEndColumnErrorf(pos, "expected named type but got %s", tnode))
		return optional.None[*schema.Ref]()
	}
	nodePath := named.Obj().Pkg().Path()
	if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) {
		destModule, ok := ftlModuleFromGoModule(nodePath).Get()
		if !ok {
			pctx.errors.add(tokenErrorf(pos, nodePath, "struct declared in non-FTL module %s", nodePath))
			return optional.None[*schema.Ref]()
		}
		dataRef := &schema.Ref{
			Pos:    goPosToSchemaPos(pos),
			Module: destModule,
			Name:   named.Obj().Name(),
		}
		for i := range named.TypeArgs().Len() {
			if typeArg, ok := visitType(pctx, pos, named.TypeArgs().At(i), isExported).Get(); ok {
				// Fully qualify the Ref if needed
				if ref, okArg := typeArg.(*schema.Ref); okArg {
					if ref.Module == "" {
						ref.Module = destModule
					}
					typeArg = ref
				}
				dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
			}
		}
		return optional.Some[*schema.Ref](dataRef)
	}

	out := &schema.Data{
		Pos:    goPosToSchemaPos(pos),
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
		Export: isExported,
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
		if typeArg, ok := visitType(pctx, pos, named.TypeArgs().At(i), isExported).Get(); ok {
			dataRef.TypeParameters = append(dataRef.TypeParameters, typeArg)
		}
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
		pctx.errors.add(tokenErrorf(pos, named.String(), "expected struct but got %s", named))
		return optional.None[*schema.Ref]()
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

func visitConst(pctx *parseContext, cnode *types.Const) optional.Option[schema.Value] {
	if b, ok := cnode.Type().Underlying().(*types.Basic); ok {
		switch b.Kind() {
		case types.String:
			value, err := strconv.Unquote(cnode.Val().String())
			if err != nil {
				pctx.errors.add(tokenWrapf(cnode.Pos(), cnode.Val().String(), err, "unsupported string constant"))
				return optional.None[schema.Value]()
			}
			return optional.Some[schema.Value](&schema.StringValue{Pos: goPosToSchemaPos(cnode.Pos()), Value: value})

		case types.Int, types.Int64:
			value, err := strconv.ParseInt(cnode.Val().String(), 10, 64)
			if err != nil {
				pctx.errors.add(tokenWrapf(cnode.Pos(), cnode.Val().String(), err, "unsupported int constant"))
				return optional.None[schema.Value]()
			}
			return optional.Some[schema.Value](&schema.IntValue{Pos: goPosToSchemaPos(cnode.Pos()), Value: int(value)})
		default:
			return optional.None[schema.Value]()
		}
	}
	return optional.None[schema.Value]()
}

func visitType(pctx *parseContext, pos token.Pos, tnode types.Type, isExported bool) optional.Option[schema.Type] {
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return optional.Some[schema.Type](&schema.Ref{Pos: goPosToSchemaPos(pos), Name: tparam.Obj().Id()})
	}
	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			if _, ok := visitType(pctx, pos, named.Underlying(), isExported).Get(); !ok {
				return optional.None[schema.Type]()
			}
			nodePath := named.Obj().Pkg().Path()
			if pctx.enums[named.Obj().Name()] != nil {
				return optional.Some[schema.Type](&schema.Ref{
					Pos:    goPosToSchemaPos(pos),
					Module: pctx.module.Name,
					Name:   named.Obj().Name(),
				})
			} else if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) {
				// If this type is named and declared in another module, it's a reference.
				// The only basic-typed references supported are enums.
				if !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
					pctx.errors.add(noEndColumnErrorf(pos,
						"unsupported external type %q", named.Obj().Pkg().Path()+"."+named.Obj().Name()))
					return optional.None[schema.Type]()
				}
				base := path.Dir(pctx.pkg.PkgPath)
				destModule := path.Base(strings.TrimPrefix(nodePath, base+"/"))
				enumRef := &schema.Ref{
					Pos:    goPosToSchemaPos(pos),
					Module: destModule,
					Name:   named.Obj().Name(),
				}
				return optional.Some[schema.Type](enumRef)
			}
		}

		switch underlying.Kind() {
		case types.String:
			return optional.Some[schema.Type](&schema.String{Pos: goPosToSchemaPos(pos)})

		case types.Int, types.Int64:
			return optional.Some[schema.Type](&schema.Int{Pos: goPosToSchemaPos(pos)})

		case types.Bool:
			return optional.Some[schema.Type](&schema.Bool{Pos: goPosToSchemaPos(pos)})

		case types.Float64:
			return optional.Some[schema.Type](&schema.Float{Pos: goPosToSchemaPos(pos)})

		default:
			return optional.None[schema.Type]()
		}

	case *types.Struct:
		named, ok := tnode.(*types.Named)
		if !ok {
			if ref, ok := visitStruct(pctx, pos, tnode, isExported).Get(); ok {
				return optional.Some[schema.Type](ref)
			}
			return optional.None[schema.Type]()
		}

		// Special-cased types.
		switch named.Obj().Pkg().Path() + "." + named.Obj().Name() {
		case "time.Time":
			return optional.Some[schema.Type](&schema.Time{Pos: goPosToSchemaPos(pos)})

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Unit":
			return optional.Some[schema.Type](&schema.Unit{Pos: goPosToSchemaPos(pos)})

		case "github.com/TBD54566975/ftl/go-runtime/ftl.Option":
			if underlying, ok := visitType(pctx, pos, named.TypeArgs().At(0), isExported).Get(); ok {
				return optional.Some[schema.Type](&schema.Optional{Pos: goPosToSchemaPos(pos), Type: underlying})
			}
			return optional.None[schema.Type]()

		default:
			nodePath := named.Obj().Pkg().Path()
			if !strings.HasPrefix(nodePath, pctx.pkg.PkgPath) && !strings.HasPrefix(nodePath, "ftl/") {
				pctx.errors.add(noEndColumnErrorf(pos, "unsupported external type %s", nodePath+"."+named.Obj().Name()))
				return optional.None[schema.Type]()
			}
			if ref, ok := visitStruct(pctx, pos, tnode, isExported).Get(); ok {
				return optional.Some[schema.Type](ref)
			}
			return optional.None[schema.Type]()
		}

	case *types.Map:
		return visitMap(pctx, pos, underlying, isExported)

	case *types.Slice:
		return visitSlice(pctx, pos, underlying, isExported)

	case *types.Interface:
		if underlying.String() == "any" {
			return optional.Some[schema.Type](&schema.Any{Pos: goPosToSchemaPos(pos)})
		}
		return optional.None[schema.Type]()

	default:
		return optional.None[schema.Type]()
	}
}

func visitMap(pctx *parseContext, pos token.Pos, tnode *types.Map, isExported bool) optional.Option[schema.Type] {
	key, ok := visitType(pctx, pos, tnode.Key(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	value, ok := visitType(pctx, pos, tnode.Elem(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Map{
		Pos:   goPosToSchemaPos(pos),
		Key:   key,
		Value: value,
	})
}

func visitSlice(pctx *parseContext, pos token.Pos, tnode *types.Slice, isExported bool) optional.Option[schema.Type] {
	// If it's a []byte, treat it as a Bytes type.
	if basic, ok := tnode.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.Byte {
		return optional.Some[schema.Type](&schema.Bytes{Pos: goPosToSchemaPos(pos)})
	}
	value, ok := visitType(pctx, pos, tnode.Elem(), isExported).Get()
	if !ok {
		return optional.None[schema.Type]()
	}

	return optional.Some[schema.Type](&schema.Array{
		Pos:     goPosToSchemaPos(pos),
		Element: value,
	})
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
	pkg            *packages.Package
	pkgs           []*packages.Package
	module         *schema.Module
	nativeNames    NativeNames
	enums          enums
	enumInterfaces enumInterfaces
	activeVerb     *schema.Verb
	errors         errorSet
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
