package analyzers

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/TBD54566975/golang-tools/go/analysis/passes/inspect"
	"github.com/TBD54566975/golang-tools/go/ast/inspector"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
)

// TypeAliasExtractor extracts type aliases to the module schema.
var TypeAliasExtractor = &analysis.Analyzer{
	Name:             "typealias",
	Doc:              "extracts type aliases to the module schema",
	Run:              extractTypeAliases,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	ResultType:       reflect.TypeFor[result](),
	RunDespiteErrors: true,
}

func extractTypeAliases(pass *analysis.Pass) (interface{}, error) {
	nn := NativeNames{}
	aliases := []schema.Decl{}
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.GenDecl) //nolint:forcetypeassert
		directives := parseDirectives(pass, node, node.Doc)
		for _, dir := range directives {
			if len(node.Specs) != 1 {
				pass.Report(errorf(node, "error parsing ftl directive: expected exactly one type declaration"))
				return
			}
			t, ok := node.Specs[0].(*ast.TypeSpec)
			if !ok {
				return
			}

			aType := pass.Pkg.Scope().Lookup(t.Name.Name)
			nativeName := aType.Pkg().Name() + "." + aType.Name()
			te, ok := dir.(*directiveTypeAlias)
			if !ok {
				continue
			}
			if len(directives) > 1 {
				pass.Report(errorf(node, "only one directive expected for type alias"))
			}

			var sType optional.Option[schema.Type]
			typ := pass.TypesInfo.TypeOf(t.Type)
			if named, ok := typ.(*types.Named); ok {
				sType = extractRef(pass, node.Pos(), named, te.IsExported())
			} else {
				sType = extractType(pass, node.Pos(), typ, te.IsExported())
			}

			if !sType.Ok() {
				pass.Report(errorf(node, "could not extract type for type alias"))
				return
			}

			alias := &schema.TypeAlias{
				Pos:      goPosToSchemaPos(pass.Fset, node.Pos()),
				Comments: extractComments(node.Doc),
				Name:     strcase.ToUpperCamel(t.Name.Name),
				Export:   te.IsExported(),
				Type:     sType.MustGet(),
			}
			nn[alias] = nativeName
			aliases = append(aliases, alias)
		}
	})
	return result{
		decls:       aliases,
		nativeNames: nn,
	}, nil
}
