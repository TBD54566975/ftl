package fsm

import (
	"go/ast"
	"go/types"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
)

const (
	ftlFSMFuncPath        = "github.com/TBD54566975/ftl/go-runtime/ftl.FSM"
	ftlTransitionFuncPath = "github.com/TBD54566975/ftl/go-runtime/ftl.Transition"
	ftlStartFuncPath      = "github.com/TBD54566975/ftl/go-runtime/ftl.Start"
)

// Extractor extracts FSMs.
var Extractor = common.NewCallDeclExtractor[*schema.FSM]("fsm", Extract, ftlFSMFuncPath)

func Extract(pass *analysis.Pass, obj types.Object, node *ast.GenDecl, callExpr *ast.CallExpr, callPath string) optional.Option[*schema.FSM] {
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
