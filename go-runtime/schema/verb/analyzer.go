package verb

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/ftl/go-runtime/schema/initialize"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

type resourceType int

const (
	none resourceType = iota
	verbClient
	databaseHandle
	mappedHandle
)

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

			// if there is no resolved resource, this must either be the context or request parameter:
			// Verb(context.Context, <request>, <resource1>, <resource2>, ...)
			if r == nil {
				if idx > 1 {
					common.Errorf(pass, param, "unsupported verb parameter type; verbs must have the "+
						"signature func(Context, Request?, Resources...)")
				}
				if idx == 1 {
					hasRequest = true
				}
				continue
			}

			if r.ref != nil {
				rt, err := r.toMetadataType()
				if err != nil {
					common.Wrapf(pass, param, err, "")
				}
				orderedResourceParams = append(orderedResourceParams, common.VerbResourceParam{
					Ref:    r.ref,
					Type:   rt,
					Mapper: r.mapper,
				})
			}
			switch r.typ {
			case verbClient:
				paramObj := common.GetObjectForNode(pass.TypesInfo, param.Type).MustGet()
				common.MarkIncludeNativeName(pass, paramObj, r.ref)
				verb.AddCall(r.ref)
			case databaseHandle:
				verb.AddDatabase(r.ref)
			case mappedHandle:
				common.Errorf(pass, param, "could not resolve underlying resource for mapped handle")
			case none:
				common.Errorf(pass, param, "unsupported verb parameter type; verbs must have the "+
					"signature func(Context, Request?, Resources...)")
			}
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
	verb.SortMetadata()
	common.MarkVerbResourceParamOrder(pass, obj, orderedResourceParams)
	return optional.Some(verb)
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

type resource struct {
	ref    *schema.Ref
	typ    resourceType
	mapper optional.Option[types.Object]
}

func (r resource) toMetadataType() (schema.Metadata, error) {
	switch r.typ {
	case verbClient:
		return &schema.MetadataCalls{}, nil
	case databaseHandle:
		return &schema.MetadataDatabases{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type")
	}
}

func resolveResource(pass *analysis.Pass, typ ast.Expr) (*resource, error) {
	obj := common.GetObjectForNode(pass.TypesInfo, typ)
	if o, ok := obj.Get(); ok {
		if _, ok := o.Type().(*types.Alias); ok {
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

	var ref *schema.Ref
	rType := getParamResourceType(pass, obj)
	switch rType {
	case none:
		return nil, nil
	case verbClient:
		o, ok := obj.Get()
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		calleeRef, err := getResourceRef(o)
		if err != nil {
			return nil, err
		}
		calleeRef.Name = strings.TrimSuffix(calleeRef.Name, "Client")
		ref = calleeRef
	case databaseHandle:
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
		r, err := getResourceRef(idxObj)
		if err != nil {
			return nil, err
		}
		r.Name = db.Name
		ref = r
	case mappedHandle:
		idxListExpr, ok := typ.(*ast.IndexListExpr)
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type; expected ftl.MappedHandle[ResourceMapper, From, To]")
		}

		if len(idxListExpr.Indices) != 3 {
			return nil, fmt.Errorf("unsupported verb parameter type; expected ftl.MappedHandle[ResourceMapper, From, To]")
		}

		mapper, ok := idxListExpr.Indices[0].(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		if mapper.Obj == nil {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		if mapper.Obj.Decl == nil {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		ts, ok := mapper.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return nil, fmt.Errorf("unsupported verb parameter type")
		}

		var handle ast.Expr
		var handleType resourceType
		var handleObj optional.Option[types.Object]
		for _, field := range st.Fields.List {
			handleObj = common.GetObjectForNode(pass.TypesInfo, field.Type)
			if hType := getParamResourceType(pass, handleObj); hType != none {
				if handleType != none {
					return nil, fmt.Errorf("mapper contains multiple resource types")
				}
				handleType = hType
				handle = field.Type
			}
		}
		if handleType == none {
			return nil, fmt.Errorf("resource mapper does not contain a resource type; " +
				"must be a field or embedded field")
		}

		resolved, err := resolveResource(pass, handle)
		if err != nil {
			return nil, err
		}
		resolved.mapper = common.GetObjectForNode(pass.TypesInfo, mapper)
		return resolved, nil

	}
	if ref == nil {
		return nil, fmt.Errorf("unsupported verb parameter type")
	}
	return &resource{ref: ref, typ: rType}, nil
}

func getParamResourceType(pass *analysis.Pass, maybeObj optional.Option[types.Object]) resourceType {
	obj, ok := maybeObj.Get()
	if !ok {
		return none
	}
	if obj.Pkg() == nil {
		return none
	}

	switch t := obj.Type().(type) {
	case *types.Named:
		if isDatabaseHandleType(pass, t) {
			return databaseHandle
		}

		if isMappedHandleType(t) {
			return mappedHandle
		}

		if _, ok := t.Underlying().(*types.Signature); !ok {
			return none
		}

		return verbClient
	case *types.Alias:
		named, ok := t.Rhs().(*types.Named)
		if !ok {
			return none
		}
		namedObj := optional.Some[types.Object](named.Obj())
		if named.Obj() == nil {
			namedObj = optional.None[types.Object]()
		}
		return getParamResourceType(pass, namedObj)

	default:
		return none
	}
}

func getResourceRef(paramObj types.Object) (*schema.Ref, error) {
	paramModule, err := common.FtlModuleFromGoPackage(paramObj.Pkg().Path())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve module for type: %w", err)
	}
	dbRef := &schema.Ref{
		Module: paramModule,
		Name:   strcase.ToLowerCamel(paramObj.Name()),
	}
	return dbRef, nil
}

func isDatabaseHandleType(pass *analysis.Pass, named *types.Named) bool {
	if named.Obj().Pkg().Path()+"."+named.Obj().Name() != "github.com/TBD54566975/ftl/go-runtime/ftl.DatabaseHandle" {
		return false
	}

	if named.TypeParams().Len() != 1 {
		return false
	}
	typeArg := named.TypeParams().At(0)

	// type argument implements `DatabaseConfig`, e.g. DatabaseHandle[MyConfig] where MyConfig implements DatabaseConfig
	return common.IsDatabaseConfigType(pass, typeArg)
}

func isMappedHandleType(named *types.Named) bool {
	if named.Obj().Pkg().Path()+"."+named.Obj().Name() != "github.com/TBD54566975/ftl/go-runtime/ftl.MappedHandle" {
		return false
	}
	if named.TypeParams().Len() != 3 {
		return false
	}
	return true
}
