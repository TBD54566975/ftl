package compile

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal/goast"
	"github.com/TBD54566975/golang-tools/go/packages"
)

var (
	fset = token.NewFileSet()

	ftlFSMFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.FSM"
	ftlTransitionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Transition"
	ftlStartFuncPath      = "github.com/TBD54566975/ftl/go-runtime/ftl.Start"
)

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Node]string

func unexpectedDirectiveErrorf(dir directive, format string, args ...interface{}) *schema.Error {
	return schema.Errorf(dir.GetPosition(), 0, format, args...)
}

func errorf(node ast.Node, format string, args ...interface{}) *schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Errorf(pos, endCol, format, args...)
}

//nolint:unparam
func wrapf(node ast.Node, err error, format string, args ...interface{}) *schema.Error {
	pos, endCol := goNodePosToSchemaPos(node)
	return schema.Wrapf(pos, endCol, err, format, args...)
}

type errorSet map[string]*schema.Error

func (e errorSet) add(err *schema.Error) {
	e[err.Error()] = err
}

func legacyExtractModuleSchema(dir string, sch *schema.Schema, out *extract.Result) error {
	pkgs, err := packages.Load(&packages.Config{
		Dir:  dir,
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}, "./...")
	if err != nil {
		return err
	}
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages found in %q, does \"go mod tidy\" need to be run?", dir)
	}

	for _, pkg := range pkgs {
		if len(strings.Split(pkg.PkgPath, "/")) > 2 {
			// skip subpackages of a module
			continue
		}
		pctx := newParseContext(pkg, pkgs, sch, out)
		for _, file := range pkg.Syntax {
			err := goast.Visit(file, func(stack []ast.Node, next func() error) (err error) {
				node := stack[len(stack)-1]
				switch node := node.(type) {
				case *ast.CallExpr:
					visitCallExpr(pctx, node, stack)

				default:
				}
				return next()
			})
			if err != nil {
				return err
			}
		}
		if len(pctx.errors) > 0 {
			out.Errors = append(out.Errors, maps.Values(pctx.errors)...)
		}
	}
	return nil
}

func extractStringLiteralArg(node *ast.CallExpr, argIndex int) (string, *schema.Error) {
	if argIndex >= len(node.Args) {
		return "", errorf(node, "expected string argument at index %d", argIndex)
	}

	literal, ok := node.Args[argIndex].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", errorf(node, "expected string literal for argument at index %d", argIndex)
	}

	s, err := strconv.Unquote(literal.Value)
	if err != nil {
		return "", wrapf(node, err, "")
	}
	if s == "" {
		return "", errorf(node, "expected non-empty string literal for argument at index %d", argIndex)
	}
	return s, nil
}

func visitCallExpr(pctx *parseContext, node *ast.CallExpr, stack []ast.Node) {
	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return
	}
	if fn.FullName() == ftlFSMFuncPath {
		parseFSMDecl(pctx, node, stack)
	}
}

func parseVerbRef(pctx *parseContext, node ast.Expr) *schema.Ref {
	_, verbFn := deref[*types.Func](pctx.pkg, node)
	if verbFn == nil {
		return nil
	}
	moduleName, ok := ftlModuleFromGoModule(verbFn.Pkg().Path()).Get()
	if !ok {
		return nil
	}
	return &schema.Ref{
		Pos:    goPosToSchemaPos(node.Pos()),
		Module: moduleName,
		Name:   strcase.ToLowerCamel(verbFn.Name()),
	}
}

func parseFSMDecl(pctx *parseContext, node *ast.CallExpr, stack []ast.Node) {
	name, schemaErr := extractStringLiteralArg(node, 0)
	if schemaErr != nil {
		pctx.errors.add(schemaErr)
		return
	}
	if !schema.ValidateName(name) {
		pctx.errors.add(errorf(node, "FSM names must be valid identifiers"))
	}

	fsm := &schema.FSM{
		Pos:      goPosToSchemaPos(node.Pos()),
		Name:     name,
		Metadata: []schema.Metadata{},
	}
	pctx.module.Decls = append(pctx.module.Decls, fsm)

	for _, arg := range node.Args[1:] {
		call, ok := arg.(*ast.CallExpr)
		if !ok {
			pctx.errors.add(errorf(arg, "expected call to Start or Transition"))
			continue
		}
		_, fn := deref[*types.Func](pctx.pkg, call.Fun)
		if fn == nil {
			pctx.errors.add(errorf(call, "expected call to Start or Transition"))
			continue
		}
		parseFSMTransition(pctx, call, fn, fsm)
	}

	varDecl, ok := varDeclForStack(stack)
	if !ok {
		return
	}
	_, directives := commentsAndDirectivesForVar(pctx, varDecl, stack)
	for _, dir := range directives {
		if retryDir, ok := dir.(*directiveRetry); ok {
			fsm.Metadata = append(fsm.Metadata, &schema.MetadataRetry{
				Pos:        retryDir.Pos,
				Count:      retryDir.Count,
				MinBackoff: retryDir.MinBackoff,
				MaxBackoff: retryDir.MaxBackoff,
			})
		} else {
			pctx.errors.add(unexpectedDirectiveErrorf(dir, "unexpected directive %q attached for FSM", dir))
		}
	}
}

// Parse a Start or Transition call in an FSM declaration and add it to the FSM.
func parseFSMTransition(pctx *parseContext, node *ast.CallExpr, fn *types.Func, fsm *schema.FSM) {
	refs := make([]*schema.Ref, len(node.Args))
	for i, arg := range node.Args {
		ref := parseVerbRef(pctx, arg)
		if ref == nil {
			pctx.errors.add(errorf(arg, "expected a reference to a sink"))
			return
		}
		refs[i] = ref
	}
	switch fn.FullName() {
	case ftlStartFuncPath:
		if len(refs) != 1 {
			pctx.errors.add(errorf(node, "expected one reference to a sink"))
			return
		}
		fsm.Start = append(fsm.Start, refs...)

	case ftlTransitionFuncPath:
		if len(refs) != 2 {
			pctx.errors.add(errorf(node, "expected two references to sinks"))
			return
		}
		fsm.Transitions = append(fsm.Transitions, &schema.FSMTransition{
			Pos:  goPosToSchemaPos(node.Pos()),
			From: refs[0],
			To:   refs[1],
		})

	default:
		pctx.errors.add(errorf(node, "expected call to Start or Transition"))
	}
}

// varDeclForCall finds the variable being set in the stack
func varDeclForStack(stack []ast.Node) (varDecl *ast.GenDecl, ok bool) {
	for i := len(stack) - 1; i >= 0; i-- {
		if decl, ok := stack[i].(*ast.GenDecl); ok && decl.Tok == token.VAR {
			return decl, true
		}
	}
	return nil, false
}

// commentsAndDirectivesForVar extracts comments and directives from a variable declaration
func commentsAndDirectivesForVar(pctx *parseContext, variableDecl *ast.GenDecl, stack []ast.Node) (comments []string, directives []directive) {
	if variableDecl.Doc == nil {
		return []string{}, []directive{}
	}
	directives, schemaErr := parseDirectives(stack[len(stack)-1], fset, variableDecl.Doc)
	if schemaErr != nil {
		pctx.errors.add(schemaErr)
	}
	return parseComments(variableDecl.Doc), directives
}

func goPosToSchemaPos(pos token.Pos) schema.Position {
	p := fset.Position(pos)
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}
}

func goNodePosToSchemaPos(node ast.Node) (schema.Position, int) {
	p := fset.Position(node.Pos())
	return schema.Position{Filename: p.Filename, Line: p.Line, Column: p.Column, Offset: p.Offset}, fset.Position(node.End()).Column
}

func parseComments(doc *ast.CommentGroup) []string {
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
	errors      errorSet
	schema      *schema.Schema
	topicsByPos map[schema.Position]*schema.Topic
}

func newParseContext(pkg *packages.Package, pkgs []*packages.Package, sch *schema.Schema, out *extract.Result) *parseContext {
	if out.NativeNames == nil {
		out.NativeNames = NativeNames{}
	}
	return &parseContext{
		pkg:         pkg,
		pkgs:        pkgs,
		module:      out.Module,
		nativeNames: out.NativeNames,
		errors:      errorSet{},
		schema:      sch,
		topicsByPos: map[schema.Position]*schema.Topic{},
	}
}
