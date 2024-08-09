package main

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var Analyzer = &analysis.Analyzer{
	Name: "commitorrollback",
	Doc:  "Detects misues of dal.TX.CommitOrRollback",
	Run:  run,
}

// Detect that any use of dal.TX.CommitOrRollback is in a defer statement and
// takes a reference to a named error return parameter.
//
// ie. Must be in the following form
//
//	func myFunc() (err error) {
//		// ...
//		defer tx.CommitOrRollback(&err)
//	}
func run(pass *analysis.Pass) (interface{}, error) {
	var inspect func(n ast.Node) bool
	funcStack := []*ast.FuncType{}
	inspect = func(n ast.Node) bool {
		switch n := n.(type) {
		case nil:
			return false

		case *ast.FuncLit:
			funcStack = append(funcStack, n.Type)
			ast.Inspect(n.Body, inspect)
			funcStack = funcStack[:len(funcStack)-1]
			return false

		case *ast.FuncDecl:
			funcStack = append(funcStack, n.Type)
			ast.Inspect(n.Body, inspect)
			funcStack = funcStack[:len(funcStack)-1]
			return false

		case *ast.CallExpr:
			sel, ok := n.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			x, ok := sel.X.(*ast.Ident)
			if !ok || x.Name != "tx" || sel.Sel.Name != "CommitOrRollback" || len(n.Args) != 2 {
				return true
			}
			arg0, ok := n.Args[1].(*ast.UnaryExpr)
			if !ok || arg0.Op != token.AND {
				return true
			}
			arg0Ident, ok := arg0.X.(*ast.Ident)
			if !ok {
				return true
			}
			funcDecl := funcStack[len(funcStack)-1]
			funcPos := pass.Fset.Position(funcDecl.Func)
			if funcDecl.Results == nil {
				pass.Reportf(arg0.Pos(), "defer tx.CommitOrRollback(ctx, &err) should be deferred with a named error return parameter but the function at %s has no named return parameters", funcPos)
				return true
			}
			for _, field := range funcDecl.Results.List {
				if result, ok := field.Type.(*ast.Ident); ok && result.Name == "error" {
					if len(field.Names) == 0 {
						pass.Reportf(arg0.Pos(), "defer tx.CommitOrRollback(ctx, &err) should be deferred with a reference to a named error return parameter, but the function at %s has no named return parameters", funcPos)
					}
					for _, name := range field.Names {
						if name.Name != arg0Ident.Name {
							namePos := pass.Fset.Position(name.NamePos)
							pass.Reportf(arg0.Pos(), "defer tx.CommitOrRollback(&ctx, %s) should be deferred with the named error return parameter here %s", arg0Ident.Name, namePos)
						}
					}
				}
			}
		}
		return true
	}
	for _, file := range pass.Files {
		funcStack = []*ast.FuncType{}
		ast.Inspect(file, inspect)
	}
	return nil, nil
}

func main() {
	singlechecker.Main(Analyzer)
}
