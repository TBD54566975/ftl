package resourceconfig

import (
	"go/ast"
	"go/types"
	"strconv"

	"github.com/block/ftl-golang-tools/go/analysis"
	"github.com/block/ftl-golang-tools/go/analysis/passes/inspect"
	"github.com/block/ftl-golang-tools/go/ast/inspector"

	"github.com/block/ftl/go-runtime/schema/common"
)

// Extractor extracts config values relating to another decl, e.g. database configurations associated with a
// database decl.
//
// Configs follow a pattern where they implement an interface, like `ftl.DatabaseConfig`.
// We extract values by looking at known receiver methods. For example:
//
//	type FooConfig struct{}
//
//	func (f FooConfig) Name() string {
//	    return "foo"
//	}
//
// From this, we'd extract the "foo" value as the database name for `FooConfig`.
var Extractor = common.NewExtractor("resourceconfig", (*Fact)(nil), Extract)

type Tag struct{} // Tag uniquely identifies the fact type for this extractor.
type Fact = common.DefaultFact[Tag]

func Extract(pass *analysis.Pass) (interface{}, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector) //nolint:forcetypeassert
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	in.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl) //nolint:forcetypeassert

		// skip if there is no receiver
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}

		// handle both pointer and non-pointer receivers
		var ident *ast.Ident
		var ok bool
		switch expr := fn.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			ident, ok = expr.X.(*ast.Ident)
			if !ok {
				return
			}
		case *ast.Ident:
			ident = expr
		default:
			return
		}

		recType := pass.TypesInfo.TypeOf(ident)
		if recType == nil {
			return
		}

		// receiver implements ftl.DatabaseConfig
		if common.IsDatabaseConfigType(pass, recType) {
			extractDatabaseConfig(pass, getDBType(pass, ident, recType), fn, ident)
		}

	})
	return common.NewExtractorResult(pass), nil
}

func extractDatabaseConfig(pass *analysis.Pass, dbType common.DatabaseType, fn *ast.FuncDecl, receiver *ast.Ident) {
	obj, ok := common.GetObjectForNode(pass.TypesInfo, receiver).Get()
	if !ok {
		return
	}
	if len(fn.Body.List) == 0 {
		return
	}
	returnStmt, ok := fn.Body.List[0].(*ast.ReturnStmt)
	if !ok || returnStmt.Results == nil || len(returnStmt.Results) == 0 {
		return
	}

	switch fn.Name.Name {
	case "Name":
		lit, ok := returnStmt.Results[0].(*ast.BasicLit)
		if !ok {
			common.Errorf(pass, fn, "unexpected return type; must implement ftl.DatabaseConfig")
			return
		}
		name, err := strconv.Unquote(lit.Value)
		if err != nil {
			common.Errorf(pass, fn, "unexpected return type; must implement ftl.DatabaseConfig")
			return
		}
		common.MarkDatabaseConfig(pass, obj, dbType, common.DatabaseConfigMethodName, name)

	default:
		return
	}
}

func getDBType(pass *analysis.Pass, receiver *ast.Ident, receiverType types.Type) common.DatabaseType {
	if common.IsPostgresDatabaseConfigType(pass, receiverType) {
		return common.DatabaseTypePostgres
	}
	if common.IsMysqlDatabaseConfigType(pass, receiverType) {
		return common.DatabaseTypeMySQL
	}
	common.Errorf(pass, receiver, "unsupported database type %s", receiverType.String())
	return ""
}
