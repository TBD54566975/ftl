package compile

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/TBD54566975/ftl/go-runtime/schema/analyzers"
	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/internal/goast"
	"github.com/TBD54566975/golang-tools/go/ast/astutil"
	"github.com/TBD54566975/golang-tools/go/packages"
)

var (
	fset             = token.NewFileSet()
	contextIfaceType = once(func() *types.Interface {
		return mustLoadRef("context", "Context").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})
	errorIFaceType = once(func() *types.Interface {
		return mustLoadRef("builtin", "error").Type().Underlying().(*types.Interface) //nolint:forcetypeassert
	})

	ftlPkgPath              = "github.com/TBD54566975/ftl/go-runtime/ftl"
	ftlCallFuncPath         = "github.com/TBD54566975/ftl/go-runtime/ftl.Call"
	ftlFSMFuncPath          = "github.com/TBD54566975/ftl/go-runtime/ftl.FSM"
	ftlTransitionFuncPath   = "github.com/TBD54566975/ftl/go-runtime/ftl.Transition"
	ftlStartFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.Start"
	ftlConfigFuncPath       = "github.com/TBD54566975/ftl/go-runtime/ftl.Config"
	ftlSecretFuncPath       = "github.com/TBD54566975/ftl/go-runtime/ftl.Secret" //nolint:gosec
	ftlPostgresDBFuncPath   = "github.com/TBD54566975/ftl/go-runtime/ftl.PostgresDatabase"
	ftlUnitTypePath         = "github.com/TBD54566975/ftl/go-runtime/ftl.Unit"
	ftlOptionTypePath       = "github.com/TBD54566975/ftl/go-runtime/ftl.Option"
	ftlTopicFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.Topic"
	ftlSubscriptionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Subscription"
	ftlTopicHandleTypeName  = "TopicHandle"
	aliasFieldTag           = "json"
)

// NativeNames is a map of top-level declarations to their native Go names.
type NativeNames map[schema.Node]string

// enumInterfaces is a map of type enum names to the interface that variants must conform to.
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

//nolint:unparam
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

func legacyExtractModuleSchema(dir string, sch *schema.Schema, out *analyzers.ExtractResult) error {
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

				case *ast.File:
					visitFile(pctx, node)

				case *ast.FuncDecl:
					visitFuncDecl(pctx, node)

				case *ast.GenDecl:
					visitGenDecl(pctx, node)

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

// extractInitialDecls traverses the package's AST and extracts declarations needed up front (type aliases, enums and topics)
//
// This allows us to know if a type is a type alias or an enum regardless of ordering when visiting each ast node.
// - The decls get added to the pctx's module, nativeNames and enumInterfaces.
// - We only want to do a simple pass, so we do not resolve references to other types. This means the TypeAlias and Enum decls have Type = nil
//   - This get's filled in with the next pass
//
// It also helps with topics because we need to know the stack when visiting a topic decl, but the subscription may occur first.
// In this case there is no way for the subscription to make the topic exported.
func extractInitialDecls(pctx *parseContext) error {
	for _, file := range pctx.pkg.Syntax {
		err := goast.Visit(file, func(stack []ast.Node, next func() error) (err error) {
			switch node := stack[len(stack)-1].(type) {
			case *ast.GenDecl:
				if node.Tok == token.TYPE {
					extractTypeDecl(pctx, node)
				}

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

func extractTypeDecl(pctx *parseContext, node *ast.GenDecl) {
	directives, parseErr := parseDirectives(node, fset, node.Doc)
	if parseErr != nil {
		// errors collected when visiting all nodes in the next pass
		return
	}

	foundDeclType := optional.None[string]()
	for _, dir := range directives {
		if len(node.Specs) != 1 {
			// errors handled on next pass
			continue
		}
		t, ok := node.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}

		aType := pctx.pkg.Types.Scope().Lookup(t.Name.Name)
		nativeName := aType.Pkg().Name() + "." + aType.Name()

		switch dir := dir.(type) {
		case *directiveEnum:
			typ := pctx.pkg.TypesInfo.TypeOf(t.Type)
			switch underlying := typ.Underlying().(type) {
			case *types.Basic:
				enum := &schema.Enum{
					Pos:      goPosToSchemaPos(node.Pos()),
					Comments: parseComments(node.Doc),
					Name:     strcase.ToUpperCamel(t.Name.Name),
					Type:     nil, // nil until next pass, when we can visit the full type graph
					Export:   dir.IsExported(),
				}
				pctx.module.Decls = append(pctx.module.Decls, enum)
				pctx.nativeNames[enum] = nativeName
			case *types.Interface:
				if underlying.NumMethods() == 0 {
					pctx.errors.add(errorf(node, "enum discriminator %q must define at least one method", t.Name.Name))
					break
				}

				hasExportedMethod := false
				for i, n := 0, underlying.NumMethods(); i < n; i++ {
					if underlying.Method(i).Exported() {
						pctx.errors.add(noEndColumnErrorf(underlying.Method(i).Pos(), "enum discriminator %q cannot "+
							"contain exported methods", t.Name.Name))
						hasExportedMethod = true
					}
				}
				if hasExportedMethod {
					break
				}

				enum := &schema.Enum{
					Pos:      goPosToSchemaPos(node.Pos()),
					Comments: parseComments(node.Doc),
					Name:     strcase.ToUpperCamel(t.Name.Name),
					Export:   dir.IsExported(),
				}
				if iTyp, ok := typ.(*types.Interface); ok {
					pctx.nativeNames[enum] = nativeName
					pctx.module.Decls = append(pctx.module.Decls, enum)
					pctx.enumInterfaces[t.Name.Name] = iTyp
				} else {
					pctx.errors.add(errorf(node, "expected interface for type enum but got %q", typ))
				}
			}
			foundDeclType = optional.Some("enum")
		case *directiveTypeAlias, *directiveData, *directiveIngress, *directiveVerb, *directiveCronJob, *directiveRetry, *directiveExport, *directiveSubscriber:
			continue
		}
		if foundDeclType, ok := foundDeclType.Get(); ok {
			if len(directives) > 1 {
				pctx.errors.add(errorf(node, "only one directive expected for %v", foundDeclType))
			}
			break
		}
	}
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
			pctx.errors.add(errorf(node, "unexpected directive attached for topic: %T", dir))
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
	validateCallExpr(pctx, node)

	_, fn := deref[*types.Func](pctx.pkg, node.Fun)
	if fn == nil {
		return
	}
	switch fn.FullName() {
	case ftlCallFuncPath:
		parseCall(pctx, node, stack)

	case ftlConfigFuncPath, ftlSecretFuncPath:
		// Secret/config declaration: ftl.Config[<type>](<name>)
		parseConfigDecl(pctx, node, fn)

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

// validateCallExpr validates all function calls
// checks if the function call is:
// - a direct verb call to an external module
// - a direct publish call on an external module's topic
func validateCallExpr(pctx *parseContext, node *ast.CallExpr) {
	selExpr, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	var lhsIdent *ast.Ident
	if expr, ok := selExpr.X.(*ast.SelectorExpr); ok {
		lhsIdent = expr.Sel
	} else if ident, ok := selExpr.X.(*ast.Ident); ok {
		lhsIdent = ident
	} else {
		return
	}
	lhsObject := pctx.pkg.TypesInfo.ObjectOf(lhsIdent)
	if lhsObject == nil {
		return
	}
	var lhsPkgPath string
	if pkgName, ok := lhsObject.(*types.PkgName); ok {
		lhsPkgPath = pkgName.Imported().Path()
	} else {
		lhsPkgPath = lhsObject.Pkg().Path()
	}
	var lhsIsExternal bool
	if !pctx.isPathInPkg(lhsPkgPath) {
		lhsIsExternal = true
	}

	if lhsType, ok := pctx.pkg.TypesInfo.TypeOf(selExpr.X).(*types.Named); ok && lhsType.Obj().Pkg() != nil && lhsType.Obj().Pkg().Path() == ftlPkgPath {
		// Calling a function on an FTL type
		if lhsType.Obj().Name() == ftlTopicHandleTypeName && selExpr.Sel.Name == "Publish" {
			if lhsIsExternal {
				pctx.errors.add(errorf(node, "can not publish directly to topics in other modules"))
			}
		}
		return
	}

	if lhsIsExternal && strings.HasPrefix(lhsPkgPath, "ftl/") {
		if sig, ok := pctx.pkg.TypesInfo.TypeOf(selExpr.Sel).(*types.Signature); ok && sig.Recv() == nil {
			// can not call functions in external modules directly
			pctx.errors.add(errorf(node, "can not call verbs in other modules directly: use ftl.Call(…) instead"))
		}
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
			pctx.errors.add(errorf(node, "unexpected directive attached for FSM: %T", dir))
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

func parseConfigDecl(pctx *parseContext, node *ast.CallExpr, fn *types.Func) {
	name, schemaErr := extractStringLiteralArg(node, 0)
	if schemaErr != nil {
		pctx.errors.add(schemaErr)
		return
	}
	if !schema.ValidateName(name) {
		pctx.errors.add(errorf(node, "config and secret names must be valid identifiers"))
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

func visitFile(pctx *parseContext, node *ast.File) {
	if node.Doc == nil {
		return
	}
	pctx.module.Comments = parseComments(node.Doc)
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

func visitGenDecl(pctx *parseContext, node *ast.GenDecl) {
	switch node.Tok {
	case token.TYPE:
		directives, err := parseDirectives(node, fset, node.Doc)
		if err != nil {
			pctx.errors.add(err)
		}
		maybeVisitTypeEnumVariant(pctx, node, directives)

		if node.Doc == nil {
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
				if t, ok := node.Specs[0].(*ast.TypeSpec); ok {
					isExported := false
					if exportableDir, ok := dir.(exportable); ok {
						isExported = exportableDir.IsExported()
					}
					// We have already collected enum and type alias declarations in extractTypeDecls
					// On this second pass we can visit deeper and pull out the type information
					typ := pctx.pkg.TypesInfo.TypeOf(t.Type)
					if _, ok := dir.(*directiveEnum); ok {
						enumOption, enumInterface := pctx.getEnumForTypeName(t.Name.Name)
						enum, ok := enumOption.Get()
						if !ok {
							// This case can be reached if a type is both an enum and a typealias.
							// Error is already reported in extractTypeDecls
							return
						}
						switch typ.Underlying().(type) {
						case *types.Basic:
							if sType, ok := visitType(pctx, node.Pos(), typ, isExported).Get(); ok {
								enum.Type = sType
							} else {
								pctx.errors.add(errorf(node, "unsupported type %q for value enum",
									pctx.pkg.TypesInfo.TypeOf(t.Type).Underlying()))
							}
						case *types.Interface:
							if !enumInterface.Ok() {
								pctx.errors.add(errorf(node, "could not find interface for type enum"))
							}
						}
					} else {
						visitType(pctx, node.Pos(), pctx.pkg.TypesInfo.Defs[t.Name].Type(), isExported)
					}
				}
			case *directiveIngress, *directiveCronJob, *directiveRetry, *directiveExport, *directiveSubscriber, *directiveTypeAlias:
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

func maybeVisitTypeEnumVariant(pctx *parseContext, node *ast.GenDecl, directives []directive) {
	if len(node.Specs) != 1 {
		return
	}
	// `type NAME TYPE` e.g. type Scalar string
	t, ok := node.Specs[0].(*ast.TypeSpec)
	if !ok {
		return
	}
	typ := pctx.pkg.TypesInfo.TypeOf(t.Type)
	if typeInterface, ok := typ.Underlying().(*types.Interface); ok {
		// Type enums should not count as variants of themselves
		pctx.enumInterfaces[t.Name.Name] = typeInterface
		return
	}

	enumVariant := &schema.EnumVariant{
		Pos:      goPosToSchemaPos(node.Pos()),
		Comments: parseComments(node.Doc),
		Name:     strcase.ToUpperCamel(t.Name.Name),
	}

	matchedEnumNames := []string{}

	// iterate in a predictable way to make sure we are not flipflopping between builds of which type enum counts as first
	allEnumNames := maps.Keys(pctx.enumInterfaces)
	slices.Sort(allEnumNames)
	for _, enumName := range allEnumNames {
		interfaceNode := pctx.enumInterfaces[enumName]

		// If the type declared is an enum variant, then it must implement
		// the interface of a type enum
		named, ok := pctx.pkg.Types.Scope().Lookup(t.Name.Name).Type().(*types.Named)
		if !ok {
			continue
		}
		if !types.Implements(named, interfaceNode) {
			continue
		}

		enumOption, _ := pctx.getEnumForTypeName(enumName)
		enum, ok := enumOption.Get()
		if !ok {
			pctx.errors.add(errorf(node, "could not find enum called %s", enumName))
			continue
		}

		matchedEnumNames = append(matchedEnumNames, enumName)
		if len(matchedEnumNames) > 1 {
			continue
		}

		if enum.VariantForName(enumVariant.Name).Ok() {
			continue
		}

		// If any directives on this node are exported, then the
		// enum variant node is considered exported. Also, if the
		// parent enum itself is exported, then all its variants
		// should transitively also be exported.
		isExported := enum.IsExported()
		for _, dir := range directives {
			if exportableDir, ok := dir.(exportable); ok {
				isExported = exportableDir.IsExported() || isExported
			}
		}
		vType, ok := visitTypeValue(pctx, named, t.Type, nil, isExported).Get()
		if !ok {
			pctx.errors.add(errorf(node, "unsupported type %q for type enum variant", named))
			continue
		}
		enumVariant.Value = vType
		enum.Variants = append(enum.Variants, enumVariant)
		pctx.nativeNames[enumVariant] = named.Obj().Pkg().Name() + "." + named.Obj().Name()
	}
	if len(matchedEnumNames) > 1 {
		slices.Sort(matchedEnumNames)
		pctx.errors.add(errorf(node, "type can not be a variant of more than 1 type enums (%s)", strings.Join(matchedEnumNames, ", ")))
	}
}

func visitTypeValue(pctx *parseContext, named *types.Named, tnode ast.Expr, index types.Type, isExported bool) optional.Option[*schema.TypeValue] {
	switch typ := tnode.(type) {
	// Selector expression e.g. ftl.Unit, ftl.Option, foo.Bar
	case *ast.SelectorExpr:
		var ident *ast.Ident
		var ok bool
		if ident, ok = typ.X.(*ast.Ident); !ok {
			return optional.None[*schema.TypeValue]()
		}

		for _, im := range maps.Values(pctx.pkg.Imports) {
			if im.Name != ident.Name {
				continue
			}
			switch im.ID + "." + typ.Sel.Name {
			case "time.Time":
				return optional.Some(&schema.TypeValue{
					Pos:   goPosToSchemaPos(tnode.Pos()),
					Value: &schema.Time{},
				})
			case ftlUnitTypePath:
				return optional.Some(&schema.TypeValue{
					Pos:   goPosToSchemaPos(tnode.Pos()),
					Value: &schema.Unit{},
				})
			case ftlOptionTypePath:
				if index == nil {
					return optional.None[*schema.TypeValue]()
				}

				if vt, ok := visitType(pctx, tnode.Pos(), index, isExported).Get(); ok {
					return optional.Some(&schema.TypeValue{
						Pos: goPosToSchemaPos(tnode.Pos()),
						Value: &schema.Optional{
							Pos:  goPosToSchemaPos(tnode.Pos()),
							Type: vt,
						},
					})
				}
			default: // Data ref
				externalModuleName, ok := ftlModuleFromGoModule(im.ID).Get()
				if !ok {
					pctx.errors.add(errorf(tnode, "package %q is not in the ftl namespace", im.ID))
					return optional.None[*schema.TypeValue]()
				}
				return optional.Some(&schema.TypeValue{
					Pos: goPosToSchemaPos(tnode.Pos()),
					Value: &schema.Ref{
						Pos:    goPosToSchemaPos(tnode.Pos()),
						Module: externalModuleName,
						Name:   typ.Sel.Name,
					},
				})
			}
		}

	case *ast.IndexExpr: // Generic type, e.g. ftl.Option[string]
		if se, ok := typ.X.(*ast.SelectorExpr); ok {
			return visitTypeValue(pctx, named, se, pctx.pkg.TypesInfo.TypeOf(typ.Index), isExported)
		}

	default:
		variantNode := pctx.pkg.TypesInfo.TypeOf(tnode)
		if _, ok := variantNode.(*types.Struct); ok {
			variantNode = named
		}
		if typ, ok := visitType(pctx, tnode.Pos(), variantNode, isExported).Get(); ok {
			return optional.Some(&schema.TypeValue{Value: typ})
		} else {
			pctx.errors.add(errorf(tnode, "unsupported type %q for type enum variant", named))
		}
	}

	return optional.None[*schema.TypeValue]()
}

func visitValueSpec(pctx *parseContext, node *ast.ValueSpec) {
	var enum *schema.Enum
	i, ok := node.Type.(*ast.Ident)
	if !ok {
		return
	}
	enumOption, enumInterface := pctx.getEnumForTypeName(i.Name)
	enum, ok = enumOption.Get()
	if !ok {
		return
	}
	if enumInterface.Ok() {
		pctx.errors.add(errorf(node, "cannot attach enum value to %s because it a type enum", enum.Name))
		return
	}
	maybeErrorOnInvalidEnumMixing(pctx, node, enum.Name)

	c, ok := pctx.pkg.TypesInfo.Defs[node.Names[0]].(*types.Const)
	if !ok {
		pctx.errors.add(errorf(node, "could not extract enum %s: expected exactly one variant name", enum.Name))
		return
	}

	if value, ok := visitConst(pctx, c).Get(); ok {
		variant := &schema.EnumVariant{
			Pos:      goPosToSchemaPos(c.Pos()),
			Comments: parseComments(node.Doc),
			Name:     strcase.ToUpperCamel(c.Id()),
			Value:    value,
		}
		enum.Variants = append(enum.Variants, variant)
	} else {
		pctx.errors.add(errorf(node, "unsupported type %q for enum variant %q", c.Type(), c.Name()))
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
	for _, decl := range pctx.module.Decls {
		enum, ok := decl.(*schema.Enum)
		if !ok {
			continue
		}
		for _, variant := range enum.Variants {
			if variant.Name == enumName {
				pctx.errors.add(errorf(node, "cannot attach enum value to %s because it is a variant of type enum %s, not a value enum", enumName, enum.Name))
			}
		}
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
		case *directiveRetry:
			metadata = append(metadata, &schema.MetadataRetry{
				Pos:        dir.Pos,
				Count:      dir.Count,
				MinBackoff: dir.MinBackoff,
				MaxBackoff: dir.MaxBackoff,
			})
		case *directiveSubscriber:
			isVerb = true
			metadata = append(metadata, &schema.MetadataSubscriber{
				Pos:  dir.Pos,
				Name: dir.Name,
			})
		case *directiveData, *directiveEnum, *directiveTypeAlias, *directiveExport:
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
		Comments: parseComments(node.Doc),
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
	enumInterfaces enumInterfaces
	errors         errorSet
	schema         *schema.Schema
	topicsByPos    map[schema.Position]*schema.Topic
}

func newParseContext(pkg *packages.Package, pkgs []*packages.Package, sch *schema.Schema, out *analyzers.ExtractResult) *parseContext {
	if out.NativeNames == nil {
		out.NativeNames = NativeNames{}
	}
	return &parseContext{
		pkg:            pkg,
		pkgs:           pkgs,
		module:         out.Module,
		nativeNames:    out.NativeNames,
		enumInterfaces: enumInterfaces{},
		errors:         errorSet{},
		schema:         sch,
		topicsByPos:    map[schema.Position]*schema.Topic{},
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

// getEnumForTypeName returns the enum and interface for a given type name.
func (p *parseContext) getEnumForTypeName(name string) (optional.Option[*schema.Enum], optional.Option[*types.Interface]) {
	aDecl, ok := p.getDeclForTypeName(name).Get()
	if !ok {
		return optional.None[*schema.Enum](), optional.None[*types.Interface]()
	}
	decl, ok := aDecl.(*schema.Enum)
	if !ok {
		return optional.None[*schema.Enum](), optional.None[*types.Interface]()
	}
	nativeName, ok := p.nativeNames[decl]
	if !ok {
		return optional.None[*schema.Enum](), optional.None[*types.Interface]()
	}
	enumInterface, isTypeEnum := p.enumInterfaces[strings.Split(nativeName, ".")[1]]
	if isTypeEnum {
		return optional.Some(decl), optional.Some(enumInterface)
	}
	return optional.Some(decl), optional.None[*types.Interface]()
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
	_ = schema.Visit(node, func(n schema.Node, next func() error) error {
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
