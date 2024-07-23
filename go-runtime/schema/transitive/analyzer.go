package transitive

import (
	"go/ast"
	"go/types"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/schema/common"
	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
)

// Extractor extracts transitive schema.Decls to the module schema.
//
// This extractor is used to extract schema.Decls that are implicitly included in the schema via other schema.Decls
// but not themselves explicitly annotated.
var Extractor = common.NewExtractor("transitive", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

// Extract traverses all schema type root AST nodes and determines if a node has been marked for extraction.
//
// Transitive data decls are marked via "facts", annotating the object which must be extracted to the schema with
// common.NeedsExtraction. This allows us to identify objects for extraction that are not explicitly
// annotated with an FTL directive.
func Extract(pass *analysis.Pass) (interface{}, error) {
	needsExtraction := sets.NewSet[types.Object]()
	for obj, fact := range common.GetAllFactsExtractionStatus(pass) {
		if _, ok := fact.(*common.NeedsExtraction); ok {
			needsExtraction.Add(obj)
		}
	}

	visited := sets.NewSet[types.Object]()
	for !needsExtraction.IsEmpty() {
		extractTransitive(pass, needsExtraction)
		visited.Append(needsExtraction.ToSlice()...)
		needsExtraction = refreshNeedsExtraction(pass, visited)
	}
	return common.NewExtractorResult(pass), nil
}

func refreshNeedsExtraction(pass *analysis.Pass, visited sets.Set[types.Object]) sets.Set[types.Object] {
	facts := sets.NewSet[types.Object]()
	for obj := range common.GetCurrentPassFacts[*common.NeedsExtraction](pass) {
		if visited.Contains(obj) {
			continue
		}
		if obj.Pkg().Path() == pass.Pkg.Path() {
			facts.Add(obj)
		}
	}
	return facts
}

func extractTransitive(pass *analysis.Pass, needsExtraction sets.Set[types.Object]) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
		(*ast.FuncDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		obj, ok := common.GetObjectForNode(pass.TypesInfo, n).Get()
		if !ok {
			return
		}
		if !needsExtraction.Contains(obj) {
			return
		}
		schType, ok := inferDeclType(pass, n, obj).Get()
		if !ok {
			// if we can't infer the type, try to extract it as data
			schType = &schema.Data{}
		}
		extract, err := common.ExtractFuncForDecl(schType)
		if err != nil {
			// unmigrated, skip
			// temporarily marking as extracted to avoid infinite loop
			common.MarkSchemaDecl(pass, obj, nil)
			return
		}
		if decl, ok := extract(pass, n, obj).Get(); ok {
			common.MarkSchemaDecl(pass, obj, decl)
		} else {
			common.MarkFailedExtraction(pass, obj)
		}
	})
}

func inferDeclType(pass *analysis.Pass, node ast.Node, obj types.Object) optional.Option[schema.Decl] {
	if md, ok := common.GetFactForObject[*common.ExtractedMetadata](pass, obj).Get(); ok {
		if md.Type != nil {
			return optional.Some[schema.Decl](md.Type)
		}
	}

	ts, ok := node.(*ast.TypeSpec)
	if !ok {
		return optional.None[schema.Decl]()
	}
	if _, ok := ts.Type.(*ast.InterfaceType); ok {
		return optional.Some[schema.Decl](&schema.Enum{})
	}
	// underlying type is external, try to extract as a type alias
	if o, ok := common.GetObjectForNode(pass.TypesInfo, ts.Type).Get(); ok && o.Pkg() != nil && common.IsExternalType(o.Pkg().Path()) {
		return optional.Some[schema.Decl](&schema.TypeAlias{})
	}
	t, ok := common.ExtractType(pass, ts.Type).Get()
	if !ok {
		return optional.None[schema.Decl]()
	}
	if !common.IsSelfReference(pass, obj, t) {
		// if this is a type alias and it has enum variants, infer to be a value enum
		for o := range common.GetAllFacts[*common.MaybeValueEnumVariant](pass) {
			if o.Type() == obj.Type() {
				return optional.Some[schema.Decl](&schema.Enum{})
			}
		}
		return optional.Some[schema.Decl](&schema.TypeAlias{})
	}
	return optional.Some[schema.Decl](&schema.Data{})
}
