package verb

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"github.com/alecthomas/types/optional"
	"github.com/block/ftl-golang-tools/go/analysis"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/go-runtime/schema/common"
	"github.com/block/ftl/go-runtime/schema/initialize"
)

type resource struct {
	ref *schema.Ref
	typ common.VerbResourceType
}

func (r resource) toMetadataType() (schema.Metadata, error) {
	switch r.typ {
	case common.VerbResourceTypeVerbClient:
		return &schema.MetadataCalls{}, nil
	case common.VerbResourceTypeDatabaseHandle:
		return &schema.MetadataDatabases{}, nil
	case common.VerbResourceTypeTopicHandle:
		return &schema.MetadataPublisher{}, nil
	case common.VerbResourceTypeConfig:
		return &schema.MetadataConfig{}, nil
	case common.VerbResourceTypeSecret:
		return &schema.MetadataSecrets{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type")
	}
}

// Extractor extracts verbs to the module schema.
var Extractor = common.NewDeclExtractor[*schema.Verb, *ast.FuncDecl]("verb", Extract)

func Extract(pass *analysis.Pass, node *ast.FuncDecl, obj types.Object) optional.Option[*schema.Verb] {
	verb := &schema.Verb{
		Pos:  common.GoPosToSchemaPos(pass.Fset, node.Pos()),
		Name: strcase.ToLowerCamel(node.Name.Name),
	}

	loaded := pass.ResultOf[initialize.Analyzer].(initialize.Result) //nolint:forcetypeassert

	hasRequest := false
	var orderedResourceParams []common.VerbResourceParam
	if !common.ApplyMetadata[*schema.Verb](pass, obj, func(md *common.ExtractedMetadata) {
		verb.Comments = md.Comments
		verb.Export = md.IsExported
		verb.Metadata = md.Metadata
		for idx, param := range node.Type.Params.List {
			r, err := resolveResource(pass, param.Type)
			if err != nil {
				common.Wrapf(pass, param, err, "")
				continue
			}

			// if this parameter can't be resolved to a resource, it must either be the context or request parameter:
			// Verb(context.Context, <request>, <resource1>, <resource2>, ...)
			if r == nil || r.typ == common.VerbResourceTypeNone {
				if idx > 1 {
					common.Errorf(pass, param, "unsupported verb parameter type; verbs must have the "+
						"signature func(Context, Request?, Resources...)")
					continue
				}
				if idx == 1 {
					hasRequest = true
				}
				continue
			}

			paramObj, ok := common.GetObjectForNode(pass.TypesInfo, param.Type).Get()
			if !ok {
				common.Errorf(pass, param, "unsupported verb parameter type")
				continue
			}
			common.MarkIncludeNativeName(pass, paramObj, r.ref)
			switch r.typ {
			case common.VerbResourceTypeVerbClient:
				verb.AddCall(r.ref)
			case common.VerbResourceTypeDatabaseHandle:
				verb.AddDatabase(r.ref)
			case common.VerbResourceTypeTopicHandle:
				verb.AddTopicPublish(r.ref)
			case common.VerbResourceTypeConfig:
				verb.AddConfig(r.ref)
			case common.VerbResourceTypeSecret:
				verb.AddSecret(r.ref)
			default:
				common.Errorf(pass, param, "unsupported verb parameter type; verbs must have the "+
					"signature func(Context, Request?, Resources...)")
			}
			rt, err := r.toMetadataType()
			if err != nil {
				common.Wrapf(pass, param, err, "")
			}
			orderedResourceParams = append(orderedResourceParams, common.VerbResourceParam{
				Ref:  r.ref,
				Type: rt,
			})
		}
	}) {
		return optional.None[*schema.Verb]()
	}

	fnt := obj.(*types.Func)             //nolint:forcetypeassert
	sig := fnt.Type().(*types.Signature) //nolint:forcetypeassert
	if sig.Recv() != nil {
		common.Errorf(pass, node, "ftl:verb cannot be a method")
		return optional.None[*schema.Verb]()
	}

	reqt, respt := checkSignature(pass, loaded, node, sig, hasRequest)
	req := optional.Some[schema.Type](&schema.Unit{})
	if reqt.Ok() {
		req = common.ExtractType(pass, node.Type.Params.List[1].Type)
	}

	resp := optional.Some[schema.Type](&schema.Unit{})
	if respt.Ok() {
		resp = common.ExtractType(pass, node.Type.Results.List[0].Type)
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

	common.MarkVerbResourceParamOrder(pass, obj, orderedResourceParams)
	return optional.Some(verb)
}

func resolveResource(pass *analysis.Pass, typ ast.Expr) (*resource, error) {
	obj, hasObj := common.GetObjectForNode(pass.TypesInfo, typ).Get()
	rType := common.GetVerbResourceType(pass, obj)
	var ref *schema.Ref
	switch rType {
	case common.VerbResourceTypeNone:
		return nil, nil
	case common.VerbResourceTypeVerbClient:
		calleeRef, ok := common.ExtractSimpleRefWithCasing(pass, typ, strcase.ToLowerCamel).Get()
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		calleeRef.Name = strings.TrimSuffix(calleeRef.Name, "Client")
		ref = calleeRef
	case common.VerbResourceTypeDatabaseHandle:
		// database parameter can either be supplied via an aliased type:
		//
		// type MyDB = ftl.DatabaseHandle[MyDatabaseConfig]
		// func MyVerb(ctx context.Context, db MyDB, ...) (..., error)
		//
		// or directly:
		// func MyVerb(ctx context.Context, db ftl.DatabaseHandle[MyDatabaseConfig], ...) (..., error)
		if hasObj {
			if _, ok := obj.Type().(*types.Alias); ok {
				ident, ok := typ.(*ast.Ident)
				if !ok || ident.Obj == nil || ident.Obj.Decl == nil {
					return nil, fmt.Errorf("unsupported verb parameter type")
				}
				ts, ok := ident.Obj.Decl.(*ast.TypeSpec)
				if !ok {
					return nil, fmt.Errorf("unsupported verb parameter type")
				}
				return resolveResource(pass, ts.Type)
			}
		}
		idxExpr, ok := typ.(*ast.IndexExpr)
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type; expected ftl.DatabaseHandle[Config]")
		}
		idxObj, ok := common.GetObjectForNode(pass.TypesInfo, idxExpr.Index).Get()
		if !ok {
			return nil, fmt.Errorf("unsupported database verb parameter type")
		}
		decl, ok := common.GetFactForObject[*common.ExtractedDecl](pass, idxObj).Get()
		if !ok {
			return nil, fmt.Errorf("no database found for config provided to database handle")
		}
		db, ok := decl.Decl.(*schema.Database)
		if !ok {
			return nil, fmt.Errorf("no database found for config provided to database handle")
		}
		module, err := common.FtlModuleFromGoPackage(idxObj.Pkg().Path())
		if err != nil {
			return nil, fmt.Errorf("failed to resolve module for type: %w", err)
		}
		ref = &schema.Ref{
			Module: module,
			Name:   db.Name,
		}
	case common.VerbResourceTypeTopicHandle, common.VerbResourceTypeSecret, common.VerbResourceTypeConfig:
		var ok bool
		if ref, ok = common.ExtractSimpleRefWithCasing(pass, typ, strcase.ToLowerCamel).Get(); !ok {
			return nil, fmt.Errorf("unsupported verb parameter type; expected ftl.TopicHandle[Event, PartitionMapper]")
		}
	}
	if ref == nil {
		return nil, fmt.Errorf("unsupported verb parameter type")
	}
	return &resource{ref: ref, typ: rType}, nil
}

func checkSignature(
	pass *analysis.Pass,
	loaded initialize.Result,
	node *ast.FuncDecl,
	sig *types.Signature,
	hasRequest bool,
) (req, resp optional.Option[*types.Var]) {
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
	if params.Len() == 0 {
		common.Errorf(pass, node, "first parameter must be context.Context")
	} else if !loaded.IsContextType(params.At(0).Type()) {
		common.TokenErrorf(pass, params.At(0).Pos(), params.At(0).Name(), "first parameter must be of type context.Context but is %s", params.At(0).Type())
	}

	if params.Len() >= 2 {
		if params.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, params.At(1).Pos(), params.At(1).Name(), "second parameter must not be ftl.Unit")
		}

		if hasRequest {
			req = optional.Some(params.At(1))
		}
	}

	if results.Len() > 2 {
		common.Errorf(pass, node, "must have at most two results (<type>, error)")
	}
	if results.Len() == 0 {
		common.Errorf(pass, node, "must at least return an error")
	} else if !loaded.IsStdlibErrorType(results.At(results.Len() - 1).Type()) {
		common.TokenErrorf(pass, results.At(results.Len()-1).Pos(), results.At(results.Len()-1).Name(), "must return an error but is %q", results.At(0).Type())
	}
	if results.Len() == 2 {
		if results.At(1).Type().String() == common.FtlUnitTypePath {
			common.TokenErrorf(pass, results.At(1).Pos(), results.At(1).Name(), "second result must not be ftl.Unit")
		}
		resp = optional.Some(results.At(0))
	}
	return req, resp
}
