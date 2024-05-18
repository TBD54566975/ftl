package typealias

import (
	"go/ast"
	"go/types"
	"reflect"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	helper "github.com/TBD54566975/ftl/go-runtime/compile/analysis"
)

type Result struct {
	TypeAliases []*schema.TypeAlias
	NativeNames map[schema.Node]string
	Errors      []*schema.Error
}

var Extractor = &analysis.Analyzer{
	Name:       "typealias",
	Doc:        "extracts type aliases to the module schema",
	Run:        extract,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: reflect.TypeFor[Result](),
}

func extract(pass *analysis.Pass) (interface{}, error) {
	var typeAliases []*schema.TypeAlias
	nativeNames := map[schema.Node]string{}
	scherrs := sets.NewSet[*schema.Error]()

	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}
	in.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.GenDecl) //nolint:forcetypeassert
		directives, err := compile.ParseDirectives(node, pass.Fset, node.Doc)
		if err != nil {
			scherrs.Add(err)
		}

		for _, dir := range directives {
			if len(node.Specs) != 1 {
				scherrs.Add(helper.Errorf(node, pass.Fset, "error parsing ftl directive: expected "+
					"exactly one type declaration"))
				return
			}
			t, ok := node.Specs[0].(*ast.TypeSpec)
			if !ok {
				return
			}

			aType := pass.Pkg.Scope().Lookup(t.Name.Name)
			nativeName := aType.Pkg().Name() + "." + aType.Name()
			te, ok := dir.(*compile.DirectiveTypeAlias)
			if !ok {
				continue
			}
			if len(directives) > 1 {
				scherrs.Add(helper.Errorf(node, pass.Fset, "only one directive expected for type alias"))
			}

			var sType optional.Option[schema.Type]
			var errs []*schema.Error
			typ := pass.TypesInfo.TypeOf(t.Type)
			if named, ok := typ.(*types.Named); ok {
				sType, errs = helper.ExtractRef(pass, node.Pos(), named, te.IsExported())
			} else {
				sType, errs = helper.ExtractType(pass, node.Pos(), typ, te.IsExported())
			}
			scherrs.Append(errs...)

			if !sType.Ok() {
				scherrs.Add(helper.Errorf(node, pass.Fset, "unsupported type %q for type alias", typ.Underlying()))
				return
			}

			alias := &schema.TypeAlias{
				Pos:      helper.GoPosToSchemaPos(pass.Fset, node.Pos()),
				Comments: helper.ExtractComments(node.Doc),
				Name:     strcase.ToUpperCamel(t.Name.Name),
				Export:   te.IsExported(),
				Type:     sType.MustGet(),
			}
			typeAliases = append(typeAliases, alias)
			nativeNames[alias] = nativeName
		}
	})
	return nil, nil
}
