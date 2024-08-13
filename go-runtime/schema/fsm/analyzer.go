package fsm

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
)

const (
	ftlFSMHandlePath      = "github.com/TBD54566975/ftl/go-runtime/ftl.FSMHandle"
	ftlFSMFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.FSM"
	ftlTransitionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Transition"
	ftlStartFuncPath      = "github.com/TBD54566975/ftl/go-runtime/ftl.Start"
)

// Extractor extracts FSMs.
// var Extractor = common.NewCallDeclExtractor[*schema.FSM]("fsm", Extract, ftlFSMFuncPath)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
var Extractor = common.NewExtractor("fsm", (*common.DefaultFact[Tag])(nil), runExtract())

type pairedNodes struct {
	decl *ast.GenDecl
	call *ast.CallExpr
}

func runExtract() func(pass *analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert

		pairs := []*pairedNodes{}
		unpairedDecls := map[string]*ast.GenDecl{}
		unpairedCalls := map[string]*ast.CallExpr{}
		in.WithStack([]ast.Node{&ast.AssignStmt{}, &ast.GenDecl{}}, func(node ast.Node, push bool, stack []ast.Node) bool {
			if !push {
				return true
			}
			switch node := node.(type) {
			case *ast.GenDecl:
				if len(node.Specs) != 1 {
					return true
				}
				valueSpec, ok := node.Specs[0].(*ast.ValueSpec)
				if !ok {
					return true
				}
				if len(valueSpec.Names) != 1 {
					return true
				}
				name := valueSpec.Names[0].Name
				if callExpr, ok := common.CallExprFromVar(node).Get(); ok && isFSMCreationExpression(pass, callExpr) {
					// this declaration of a fsm handle variable also instantiates the FSM
					pairs = append(pairs, &pairedNodes{decl: node, call: callExpr})
					return true
				}
				typeInfo, ok := common.GetTypeInfoForNode(node, pass.TypesInfo).Get()
				if !ok {
					return true
				}
				if typeInfo.String() != "*"+ftlFSMHandlePath {
					return true
				}
				// this declaration of a fsm handle variable does not actually instantiate the FSM
				fmt.Printf("found decl without instantiation %s: %v\n", name, node)
				unpairedDecls[name] = node
			case *ast.AssignStmt:
				// TODO check for any local var declarations
				if node.Tok != token.ASSIGN {
					return true
				}
				if len(node.Lhs) != 1 {
					return true
				}
				ident, ok := node.Lhs[0].(*ast.Ident)
				if !ok {
					return true
				}
				callExpr, ok := node.Rhs[0].(*ast.CallExpr)
				if !ok || !isFSMCreationExpression(pass, callExpr) {
					return true
				}
				fmt.Printf("found unpaired: %s %s\n", ident.Name, node.Tok)
				unpairedCalls[ident.Name] = callExpr
			}
			return true
		})
		for name, decl := range unpairedDecls {
			if call, ok := unpairedCalls[name]; ok {
				pairs = append(pairs, &pairedNodes{decl: decl, call: call})
			}
		}
		for _, pair := range pairs {
			obj, ok := common.GetObjectForNode(pass.TypesInfo, pair.decl).Get()
			if !ok {
				continue
			}
			if d, ok := Extract(pass, obj, pair.call).Get(); ok {
				fmt.Printf("extracted\n")
				common.MarkSchemaDecl(pass, obj, d)
			} else {
				common.MarkFailedExtraction(pass, obj)
			}
		}
		return common.NewExtractorResult(pass), nil
	}
}

func isFSMCreationExpression(pass *analysis.Pass, node *ast.CallExpr) bool {
	_, fn := common.Deref[*types.Func](pass, node.Fun)
	if fn == nil {
		return false
	}
	callPath := fn.FullName()
	return ftlFSMFuncPath == callPath
}

func Extract(pass *analysis.Pass, obj types.Object, callExpr *ast.CallExpr) optional.Option[*schema.FSM] {
	name := common.ExtractStringLiteralArg(pass, callExpr, 0)
	if !schema.ValidateName(name) {
		common.Errorf(pass, callExpr, "FSM names must be valid identifiers")
	}

	fsm := &schema.FSM{
		Pos:      common.GoPosToSchemaPos(pass.Fset, callExpr.Pos()),
		Name:     name,
		Metadata: []schema.Metadata{},
	}

	for _, arg := range callExpr.Args[1:] {
		call, ok := arg.(*ast.CallExpr)
		if !ok {
			common.Errorf(pass, arg, "expected call to Start or Transition")
			continue
		}
		_, fn := common.Deref[*types.Func](pass, call.Fun)
		if fn == nil {
			common.Errorf(pass, call, "expected call to Start or Transition")
			continue
		}
		parseFSMTransition(pass, call, fn, fsm)
	}

	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		for _, m := range md.Metadata {
			if _, ok := m.(*schema.MetadataRetry); !ok {
				common.Errorf(pass, callExpr, "unexpected metadata %q attached for FSM", m)
			}
		}
		fsm.Comments = md.Comments
		fsm.Metadata = md.Metadata
	}
	return optional.Some(fsm)
}

// Parse a Start or Transition call in an FSM declaration and add it to the FSM.
func parseFSMTransition(pass *analysis.Pass, node *ast.CallExpr, fn *types.Func, fsm *schema.FSM) {
	refs := make([]*schema.Ref, len(node.Args))
	for i, arg := range node.Args {
		ref := parseVerbRef(pass, arg)
		if ref == nil {
			common.Errorf(pass, arg, "expected a reference to a sink")
			return
		}
		refs[i] = ref
	}
	switch fn.FullName() {
	case ftlStartFuncPath:
		if len(refs) != 1 {
			common.Errorf(pass, node, "expected one reference to a sink")
			return
		}
		fsm.Start = append(fsm.Start, refs...)

	case ftlTransitionFuncPath:
		if len(refs) != 2 {
			common.Errorf(pass, node, "expected two references to sinks")
			return
		}
		fsm.Transitions = append(fsm.Transitions, &schema.FSMTransition{
			Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
			From: refs[0],
			To:   refs[1],
		})

	default:
		common.Errorf(pass, node, "expected call to Start or Transition")
	}
}

func parseVerbRef(pass *analysis.Pass, node ast.Expr) *schema.Ref {
	_, verbFn := common.Deref[*types.Func](pass, node)
	if verbFn == nil {
		return nil
	}
	moduleName, err := common.FtlModuleFromGoPackage(verbFn.Pkg().Path())
	if err != nil {
		return nil
	}
	return &schema.Ref{
		Pos:    common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Module: moduleName,
		Name:   strcase.ToLowerCamel(verbFn.Name()),
	}
}
