package verb

import (
	"go/ast"
	"go/types"
	"unicode"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/go-runtime/schema/initialize"
)

// Extractor extracts verbs to the module schema.
var Extractor = common.NewDeclExtractor[*schema.Verb, *ast.FuncDecl]("verb", Extract)

func Extract(pass *analysis.Pass, node *ast.FuncDecl, obj types.Object) optional.Option[*schema.Verb] {
	verb := &schema.Verb{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToLowerCamel(node.Name.Name),
	}
	if !common.ApplyMetadata[*schema.Verb](pass, obj, func(md *common.ExtractedMetadata) {
		verb.Comments = md.Comments
		verb.Export = md.IsExported
		verb.Metadata = md.Metadata
	}) {
		return optional.None[*schema.Verb]()
	}

	fnt := obj.(*types.Func)             //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature) //nolint:forcetypeassert
	if sig.Recv() != nil {
		common.Errorf(pass, node, "ftl:verb cannot be a method")
		return optional.None[*schema.Verb]()
	}

	reqt, respt := checkSignature(pass, node, sig)
	req := optional.Some[schema.Type](&schema.Unit{})
	if reqt.Ok() {
		req = common.ExtractType(pass, node.Type.Params.List[1])
	}
	resp := optional.Some[schema.Type](&schema.Unit{})
	if respt.Ok() {
		resp = common.ExtractType(pass, node.Type.Results.List[0])
	}

	params := sig.Params()
	results := sig.Results()
	reqV, ok := req.Get()
	if !ok {
		common.Errorf(pass, node.Type.Params.List[1], "unsupported request type %q", params.At(1).Type())
	}
	resV, ok := resp.Get()
	if !ok {
		common.Errorf(pass, node.Type.Results.List[0], "unsupported response type %q", results.At(0).Type())
	}
	verb.Request = reqV
	verb.Response = resV

	return optional.Some(verb)
}

func checkSignature(pass *analysis.Pass, node *ast.FuncDecl, sig *types.Signature) (req, resp optional.Option[*types.Var]) {
	if node.Name.Name == "" {
		common.Errorf(pass, node, "verb function must be named")
		return optional.None[*types.Var](), optional.None[*types.Var]()
	}
	if !unicode.IsUpper(rune(node.Name.Name[0])) {
		common.Errorf(pass, node, "verb name must be exported")
		return optional.None[*types.Var](), optional.None[*types.Var]()
	}

	params := sig.Params()
	results := sig.Results()
	if params.Len() > 2 {
		common.Errorf(pass, node, "must have at most two parameters (context.Context, struct)")
	}

	loaded := pass.ResultOf[initialize.Analyzer].(initialize.Result) //nolint:forcetypeassert
	if params.Len() == 0 {
		common.Errorf(pass, node, "first parameter must be context.Context")
	} else if !loaded.IsContextType(params.At(0).Type()) {
		common.TokenErrorf(pass, params.At(0).Pos(), params.At(0).Name(), "first parameter must be of type context.Context but is %s", params.At(0).Type())
	}

	if params.Len() == 2 {
		if params.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, params.At(1).Pos(), params.At(1).Name(), "second parameter must not be ftl.Unit")
		}

		req = optional.Some(params.At(1))
	}

	if results.Len() > 2 {
		common.Errorf(pass, node, "must have at most two results (<type>, error)")
	}
	if results.Len() == 0 {
		common.Errorf(pass, node, "must at least return an error")
	} else if !loaded.IsFtlErrorType(results.At(results.Len() - 1).Type()) {
		common.TokenErrorf(pass, results.At(results.Len()-1).Pos(), results.At(results.Len()-1).Name(), "must return an error but is %s", results.At(0).Type())
	}
	if results.Len() == 2 {
		if results.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, results.At(1).Pos(), results.At(1).Name(), "second result must not be ftl.Unit")
		}
		resp = optional.Some(results.At(0))
	}
	return req, resp
}
