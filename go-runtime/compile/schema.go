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
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	extract "github.com/TBD54566975/ftl/go-runtime/schema"
	"github.com/TBD54566975/ftl/internal/goast"
	"github.com/TBD54566975/golang-tools/go/ast/astutil"
	"github.com/TBD54566975/golang-tools/go/packages"
)

var (
	fset = token.NewFileSet()

	ftlCallFuncPath         = "github.com/TBD54566975/ftl/go-runtime/ftl.Call"
	ftlFSMFuncPath          = "github.com/TBD54566975/ftl/go-runtime/ftl.FSM"
	ftlTransitionFuncPath   = "github.com/TBD54566975/ftl/go-runtime/ftl.Transition"
	ftlStartFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.Start"
	ftlPostgresDBFuncPath   = "github.com/TBD54566975/ftl/go-runtime/ftl.PostgresDatabase"
	ftlTopicFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.Topic"
	ftlSubscriptionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Subscription"
	aliasFieldTag           = "json"
)

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Node]string

func noEndColumnErrorf(pos token.Pos, format string, args ...interface{}) *schema.Error {
	return tokenErrorf(pos, "", format, args...)
}

func unexpectedDirectiveErrorf(dir directive, format string, args ...interface{}) *schema.Error {
	return schema.Errorf(dir.GetPosition(), 0, format, args...)
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
		err := extractInitialDecls(pctx)
		if err != nil {
			return err
		}
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

// extractInitialDecls traverses the package's AST and extracts declarations needed up front (topics)
//
// We need to know the stack when visiting a topic decl, but the subscription may occur first.
// In this case there is no way for the subscription to make the topic exported.
func extractInitialDecls(pctx *parseContext) error {
	for _, file := range pctx.pkg.Syntax {
		err := goast.Visit(file, func(stack []ast.Node, next func() error) (err error) {
			switch node := stack[len(stack)-1].(type) {
			case *ast.CallExpr:
				_, fn := deref[*types.Func](pctx.pkg, node.Fun)
				if fn != nil && fn.FullName() == ftlTopicFuncPath {
					extractTopicDecl(pctx, node, stack)
				}
			}

			return next()
		})
		if err != nil {
			return err
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

// extractTopicDecl expects: _ = ftl.Topic[EventType]("name_literal")
func extractTopicDecl(pctx *parseContext, node *ast.CallExpr, stack []ast.Node) {
	name, nameErr := extractStringLiteralArg(node, 0)
	if nameErr != nil {
		pctx.errors.add(nameErr)
		return
	}

	varDecl, ok := varDeclForStack(stack)
	if !ok {
		pctx.errors.add(errorf(node, "expected topic declaration to be assigned to a variable"))
		return
	} else if len(varDecl.Specs) == 0 {
		pctx.errors.add(errorf(node, "expected topic declaration to have at least 1 spec"))
		return
	}
	topicVarPos := goPosToSchemaPos(varDecl.Specs[0].Pos())

	comments, directives := commentsAndDirectivesForVar(pctx, varDecl, stack)
	export := false
	for _, dir := range directives {
		if _, ok := dir.(*directiveExport); ok {
			export = true
		} else {
			pctx.errors.add(unexpectedDirectiveErrorf(dir, "unexpected directive %q attached for topic", dir))
		}
	}

	// Check for duplicates
	_, endCol := goNodePosToSchemaPos(node)
	for _, d := range pctx.module.Decls {
		existing, ok := d.(*schema.Topic)
		if ok && existing.Name == name {
			pctx.errors.add(errorf(node, "duplicate topic registration at %d:%d-%d", existing.Pos.Line, existing.Pos.Column, endCol))
			return
		}
	}

	topic := &schema.Topic{
		Pos:      goPosToSchemaPos(node.Pos()),
		Name:     name,
		Export:   export,
		Comments: comments,
		Event:    nil, // event is nil until we the next pass, when we can visit the full graph
	}
	pctx.module.Decls = append(pctx.module.Decls, topic)
	pctx.topicsByPos[topicVarPos] = topic
}

func visitCallExpr(pctx *parseContext, node *ast.CallExpr, stack []ast.Node) {
	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return
	}
	switch fn.FullName() {
	case ftlCallFuncPath:
		parseCall(pctx, node, stack)

	case ftlFSMFuncPath:
		parseFSMDecl(pctx, node, stack)

	case ftlPostgresDBFuncPath:
		parseDatabaseDecl(pctx, node, schema.PostgresDatabaseType)

	case ftlTopicFuncPath:
		parseTopicDecl(pctx, node)

	case ftlSubscriptionFuncPath:
		parseSubscriptionDecl(pctx, node)
	}
}

func parseCall(pctx *parseContext, node *ast.CallExpr, stack []ast.Node) {
	var activeFuncDecl *ast.FuncDecl
	for i := len(stack) - 1; i >= 0; i-- {
		if found, ok := stack[i].(*ast.FuncDecl); ok {
			activeFuncDecl = found
			break
		}
		// use element
	}
	if activeFuncDecl == nil {
		return
	}
	expectedVerbName := strcase.ToLowerCamel(activeFuncDecl.Name.Name)
	var activeVerb *schema.Verb
	for _, decl := range pctx.module.Decls {
		if aVerb, ok := decl.(*schema.Verb); ok && aVerb.Name == expectedVerbName {
			activeVerb = aVerb
			break
		}
	}
	if activeVerb == nil {
		return
	}
	if len(node.Args) != 3 {
		pctx.errors.add(errorf(node, "call must have exactly three arguments"))
		return
	}
	ref := parseVerbRef(pctx, node.Args[1])
	if ref == nil {
		var suffix string
		var ok bool
		ref, ok = parseSelectorRef(node.Args[1])
		if ok && pctx.schema.Resolve(ref).Ok() {
			suffix = ", does it need to be exported?"
		}
		if sel, ok := node.Args[1].(*ast.SelectorExpr); ok {
			pctx.errors.add(errorf(node.Args[1], "call first argument must be a function but is an unresolved reference to %s.%s%s", sel.X, sel.Sel, suffix))
		}
		pctx.errors.add(errorf(node.Args[1], "call first argument must be a function in an ftl module%s", suffix))
		return
	}
	activeVerb.AddCall(ref)
}

func parseSelectorRef(node ast.Expr) (*schema.Ref, bool) {
	sel, ok := node.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	return &schema.Ref{
		Pos:    goPosToSchemaPos(node.Pos()),
		Module: ident.Name,
		Name:   strcase.ToLowerCamel(sel.Sel.Name),
	}, true

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

func parseDatabaseDecl(pctx *parseContext, node *ast.CallExpr, dbType string) {
	name, schemaErr := extractStringLiteralArg(node, 0)
	if schemaErr != nil {
		pctx.errors.add(schemaErr)
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

// parseTopicDecl expects: _ = ftl.Topic[EventType]("name_literal")
func parseTopicDecl(pctx *parseContext, node *ast.CallExpr) {
	// already extracted topic in the initial pass of the ast graph
	// we did not do event type resolution yet, so we need to do that now
	name, nameErr := extractStringLiteralArg(node, 0)
	if nameErr != nil {
		// error already added in previous pass
		return
	}

	var topic *schema.Topic
	for _, d := range pctx.module.Decls {
		if d, ok := d.(*schema.Topic); ok && d.Name == name {
			topic = d
		}
	}

	// update topic's event type
	indexExpr, ok := node.Fun.(*ast.IndexExpr)
	if !ok {
		pctx.errors.add(errorf(node, "must have an event type as a type parameter"))
		return
	}
	typeParamType, ok := visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.TypeOf(indexExpr.Index), topic.Export).Get()
	if !ok {
		pctx.errors.add(errorf(node, "invalid event type"))
		return
	}
	topic.Event = typeParamType
}

// parseSubscriptionDecl expects: var _ = ftl.Subscription(topicHandle, "name_literal")
func parseSubscriptionDecl(pctx *parseContext, node *ast.CallExpr) {
	var name string
	var topicRef *schema.Ref
	if len(node.Args) != 2 {
		pctx.errors.add(errorf(node, "subscription registration must have a topic"))
		return
	}
	if topicIdent, ok := node.Args[0].(*ast.Ident); ok {
		// Topic is within module
		// we will find the subscription name from the string literal parameter
		object := pctx.pkg.TypesInfo.ObjectOf(topicIdent)
		topic, ok := pctx.topicsByPos[goPosToSchemaPos(object.Pos())]
		if !ok {
			pctx.errors.add(errorf(node, "could not find topic declaration for topic variable"))
			return
		}
		topicRef = &schema.Ref{
			Module: pctx.module.Name,
			Name:   topic.Name,
		}
	} else if topicSelExp, ok := node.Args[0].(*ast.SelectorExpr); ok {
		// External topic
		// we will derive subscription name from generated variable name
		moduleIdent, moduleOk := topicSelExp.X.(*ast.Ident)
		if !moduleOk {
			pctx.errors.add(errorf(node, "subscription registration must have a topic"))
			return
		}
		varName := topicSelExp.Sel.Name
		topicRef = &schema.Ref{
			Module: moduleIdent.Name,
			Name:   strings.ToLower(string(varName[0])) + varName[1:],
		}
	} else {
		pctx.errors.add(errorf(node, "subscription registration must have a topic"))
		return
	}

	// name
	var schemaErr *schema.Error
	name, schemaErr = extractStringLiteralArg(node, 1)
	if schemaErr != nil {
		pctx.errors.add(schemaErr)
		return
	}

	// Check for duplicates
	_, endCol := goNodePosToSchemaPos(node)
	for _, d := range pctx.module.Decls {
		existing, ok := d.(*schema.Subscription)
		if ok && existing.Name == name {
			pctx.errors.add(errorf(node, "duplicate subscription registration at %d:%d-%d", existing.Pos.Line, existing.Pos.Column, endCol))
			return
		}
	}

	decl := &schema.Subscription{
		Pos:   goPosToSchemaPos(node.Pos()),
		Name:  name,
		Topic: topicRef,
	}

	pctx.module.Decls = append(pctx.module.Decls, decl)
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

func visitStruct(pctx *parseContext, pos token.Pos, tnode types.Type, isExported bool) optional.Option[*schema.Ref] {
	named, ok := tnode.(*types.Named)
	if !ok {
		pctx.errors.add(noEndColumnErrorf(pos, "expected named type but got %s", tnode))
		return optional.None[*schema.Ref]()
	}
	nodePath := named.Obj().Pkg().Path()
	if !pctx.isPathInPkg(nodePath) {
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
		typeArgs := named.TypeArgs()
		if typeArgs == nil {
			continue
		}
		if typeArg, ok := visitType(pctx, pos, typeArgs.At(i), isExported).Get(); ok {
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
					out.Comments = parseComments(path.Doc)
				}
			case *ast.GenDecl:
				if path.Doc != nil {
					out.Comments = parseComments(path.Doc)
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

func visitType(pctx *parseContext, pos token.Pos, tnode types.Type, isExported bool) optional.Option[schema.Type] {
	if tparam, ok := tnode.(*types.TypeParam); ok {
		return optional.Some[schema.Type](&schema.Ref{Pos: goPosToSchemaPos(pos), Name: tparam.Obj().Id()})
	}

	if named, ok := tnode.(*types.Named); ok {
		// Handle refs to type aliases and enums, rather than the underlying type.
		decl, ok := pctx.getDeclForTypeName(named.Obj().Name()).Get()
		if ok {
			switch decl.(type) {
			case *schema.TypeAlias, *schema.Enum:
				return visitNamedRef(pctx, pos, named, isExported)
			case *schema.Data, *schema.Verb, *schema.Config, *schema.Secret, *schema.Database, *schema.FSM, *schema.Topic, *schema.Subscription:
			}
		}
	}

	switch underlying := tnode.Underlying().(type) {
	case *types.Basic:
		if named, ok := tnode.(*types.Named); ok {
			if !pctx.isPathInPkg(named.Obj().Pkg().Path()) {
				// external named types get treated as refs
				return visitNamedRef(pctx, pos, named, isExported)
			}
			// internal named types without decls are treated as basic types
		}
		switch underlying.Kind() {
		case types.String:
			return optional.Some[schema.Type](&schema.String{Pos: goPosToSchemaPos(pos)})

		case types.Int:
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
			pctx.errors.add(noEndColumnErrorf(pos, "expected named type but got %s", tnode))
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
			if !pctx.isPathInPkg(nodePath) && !strings.HasPrefix(nodePath, "ftl/") {
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
		if named, ok := tnode.(*types.Named); ok {
			return visitNamedRef(pctx, pos, named, isExported)
		}
		return optional.None[schema.Type]()

	default:
		return optional.None[schema.Type]()
	}
}

func visitNamedRef(pctx *parseContext, pos token.Pos, named *types.Named, isExported bool) optional.Option[schema.Type] {
	if named.Obj().Pkg() == nil {
		return optional.None[schema.Type]()
	}

	// Update the visibility of the reference if the referencer is exported (ensuring refs are transitively
	// exported as needed).
	if isExported {
		if decl, ok := pctx.getDeclForTypeName(named.Obj().Name()).Get(); ok {
			pctx.markAsExported(decl)
		}
	}

	nodePath := named.Obj().Pkg().Path()
	destModule := pctx.module.Name
	if !pctx.isPathInPkg(nodePath) {
		if !strings.HasPrefix(named.Obj().Pkg().Path(), "ftl/") {
			pctx.errors.add(noEndColumnErrorf(pos,
				"unsupported external type %q", named.Obj().Pkg().Path()+"."+named.Obj().Name()))
			return optional.None[schema.Type]()
		}
		base := path.Dir(pctx.pkg.PkgPath)
		destModule = path.Base(strings.TrimPrefix(nodePath, base+"/"))
	}
	ref := &schema.Ref{
		Pos:    goPosToSchemaPos(pos),
		Module: destModule,
		Name:   strcase.ToUpperCamel(named.Obj().Name()),
	}
	return optional.Some[schema.Type](ref)
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

func (p *parseContext) isPathInPkg(path string) bool {
	if path == p.pkg.PkgPath {
		return true
	}
	return strings.HasPrefix(path, p.pkg.PkgPath+"/")
}

func (p *parseContext) getDeclForTypeName(name string) optional.Option[schema.Decl] {
	for _, decl := range p.module.Decls {
		nativeName, ok := p.nativeNames[decl]
		if !ok {
			continue
		}
		if nativeName != p.pkg.Name+"."+name {
			continue
		}
		return optional.Some(decl)
	}
	return optional.None[schema.Decl]()
}

func (p *parseContext) markAsExported(node schema.Node) {
	_ = schema.Visit(node, func(n schema.Node, next func() error) error { //nolint:errcheck
		if decl, ok := n.(schema.Decl); ok {
			switch decl := decl.(type) {
			case *schema.Enum:
				decl.Export = true
			case *schema.TypeAlias:
				decl.Export = true
			case *schema.Data:
				decl.Export = true
			case *schema.Verb:
				decl.Export = true
			case *schema.Topic:
				decl.Export = true
			case *schema.Config, *schema.Secret, *schema.Database, *schema.FSM, *schema.Subscription:
				return next()
			}
		} else if r, ok := n.(*schema.Ref); ok {
			if r.Module != "" && r.Module != p.module.Name {
				return next()
			}
			for _, d := range p.module.Decls {
				switch d := d.(type) {
				case *schema.Enum, *schema.TypeAlias, *schema.Data, *schema.Topic, *schema.Subscription:
					if named, ok := d.(schema.Named); !ok || named.GetName() != r.Name {
						continue
					}
				case *schema.Verb, *schema.Config, *schema.Secret, *schema.Database, *schema.FSM:
					// does not support implicit exporting
					continue
				default:
					panic("unexpected decl type")
				}
				if exportableDecl, ok := d.(exportable); ok {
					if !exportableDecl.IsExported() {
						p.markAsExported(d)
					}
				}
			}
		}
		return next()
	})
}

func tokenFileContainsPos(f *token.File, pos token.Pos) bool {
	p := int(pos)
	base := f.Base()
	return base <= p && p < base+f.Size()
}
