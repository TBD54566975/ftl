// Package goast provides a useful visitor for the Go AST.
package goast

import (
	"fmt"
	goast "go/ast"
)

type VisitorFunc func(stack []goast.Node, next func() error) error

// Visit all nodes in the Go AST rooted at node, in depth-first order.
//
// The visitor function can call next() to continue traversal.
//
// Note that this is based on a direct copy of ast.Walk.
func Visit(node goast.Node, v VisitorFunc) error {
	return visitStack([]goast.Node{node}, v)
}

func visitStack(stack []goast.Node, v VisitorFunc) error { //nolint:maintidx
	return v(stack, func() error {
		// walk children
		// (the order of the cases matches the order
		// of the corresponding node sqltypes in ast.go)
		children := []goast.Node{}

		switch n := stack[len(stack)-1].(type) {
		// Comments and fields
		case *goast.Comment:
			// nothing to do

		case *goast.CommentGroup:
			for _, c := range n.List {
				children = append(children, c)
			}

		case *goast.Field:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			for _, c := range n.Names {
				children = append(children, c)
			}
			if n.Type != nil {
				children = append(children, n.Type)
			}
			if n.Tag != nil {
				children = append(children, n.Tag)
			}
			if n.Comment != nil {
				children = append(children, n.Comment)
			}

		case *goast.FieldList:
			for _, f := range n.List {
				children = append(children, f)
			}

		// Expressions
		case *goast.BadExpr, *goast.Ident, *goast.BasicLit:
			// nothing to do

		case *goast.Ellipsis:
			if n.Elt != nil {
				children = append(children, n.Elt)
			}

		case *goast.FuncLit:
			children = append(children, n.Type, n.Body)

		case *goast.CompositeLit:
			if n.Type != nil {
				children = append(children, n.Type)
			}
			for _, c := range n.Elts {
				children = append(children, c)
			}

		case *goast.ParenExpr:
			children = append(children, n.X)

		case *goast.SelectorExpr:
			children = append(children, n.X, n.Sel)

		case *goast.IndexExpr:
			children = append(children, n.X, n.Index)

		case *goast.IndexListExpr:
			children = append(children, n.X)
			for _, index := range n.Indices {
				children = append(children, index)
			}

		case *goast.SliceExpr:
			children = append(children, n.X)
			if n.Low != nil {
				children = append(children, n.Low)
			}
			if n.High != nil {
				children = append(children, n.High)
			}
			if n.Max != nil {
				children = append(children, n.Max)
			}

		case *goast.TypeAssertExpr:
			children = append(children, n.X)
			if n.Type != nil {
				children = append(children, n.Type)
			}

		case *goast.CallExpr:
			children = append(children, n.Fun)
			for _, c := range n.Args {
				children = append(children, c)
			}

		case *goast.StarExpr:
			children = append(children, n.X)

		case *goast.UnaryExpr:
			children = append(children, n.X)

		case *goast.BinaryExpr:
			children = append(children, n.X, n.Y)

		case *goast.KeyValueExpr:
			children = append(children, n.Key, n.Value)

		// Types
		case *goast.ArrayType:
			if n.Len != nil {
				children = append(children, n.Len)
			}
			children = append(children, n.Elt)

		case *goast.StructType:
			children = append(children, n.Fields)

		case *goast.FuncType:
			if n.TypeParams != nil {
				children = append(children, n.TypeParams)
			}
			if n.Params != nil {
				children = append(children, n.Params)
			}
			if n.Results != nil {
				children = append(children, n.Results)
			}

		case *goast.InterfaceType:
			children = append(children, n.Methods)

		case *goast.MapType:
			children = append(children, n.Key, n.Value)

		case *goast.ChanType:
			children = append(children, n.Value)

		// Statements
		case *goast.BadStmt:
			// nothing to do

		case *goast.DeclStmt:
			children = append(children, n.Decl)

		case *goast.EmptyStmt:
			// nothing to do

		case *goast.LabeledStmt:
			children = append(children, n.Label, n.Stmt)

		case *goast.ExprStmt:
			children = append(children, n.X)

		case *goast.SendStmt:
			children = append(children, n.Chan, n.Value)

		case *goast.IncDecStmt:
			children = append(children, n.X)

		case *goast.AssignStmt:
			for _, c := range n.Lhs {
				children = append(children, c)
			}
			for _, c := range n.Rhs {
				children = append(children, c)
			}

		case *goast.GoStmt:
			children = append(children, n.Call)

		case *goast.DeferStmt:
			children = append(children, n.Call)

		case *goast.ReturnStmt:
			for _, c := range n.Results {
				children = append(children, c)
			}

		case *goast.BranchStmt:
			if n.Label != nil {
				children = append(children, n.Label)
			}

		case *goast.BlockStmt:
			for _, c := range n.List {
				children = append(children, c)
			}

		case *goast.IfStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			children = append(children, n.Cond, n.Body)
			if n.Else != nil {
				children = append(children, n.Else)
			}

		case *goast.CaseClause:
			for _, c := range n.List {
				children = append(children, c)
			}
			for _, c := range n.Body {
				children = append(children, c)
			}

		case *goast.SwitchStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			if n.Tag != nil {
				children = append(children, n.Tag)
			}
			children = append(children, n.Body)

		case *goast.TypeSwitchStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			children = append(children, n.Assign, n.Body)

		case *goast.CommClause:
			if n.Comm != nil {
				children = append(children, n.Comm)
			}
			for _, c := range n.Body {
				children = append(children, c)
			}

		case *goast.SelectStmt:
			children = append(children, n.Body)

		case *goast.ForStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			if n.Cond != nil {
				children = append(children, n.Cond)
			}
			if n.Post != nil {
				children = append(children, n.Post)
			}
			children = append(children, n.Body)

		case *goast.RangeStmt:
			if n.Key != nil {
				children = append(children, n.Key)
			}
			if n.Value != nil {
				children = append(children, n.Value)
			}
			children = append(children, n.X, n.Body)

		// Declarations
		case *goast.ImportSpec:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			if n.Name != nil {
				children = append(children, n.Name)
			}
			children = append(children, n.Path)
			if n.Comment != nil {
				children = append(children, n.Comment)
			}

		case *goast.ValueSpec:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			for _, c := range n.Names {
				children = append(children, c)
			}
			if n.Type != nil {
				children = append(children, n.Type)
			}
			for _, c := range n.Values {
				children = append(children, c)
			}
			if n.Comment != nil {
				children = append(children, n.Comment)
			}

		case *goast.TypeSpec:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			children = append(children, n.Name)
			if n.TypeParams != nil {
				children = append(children, n.TypeParams)
			}
			children = append(children, n.Type)
			if n.Comment != nil {
				children = append(children, n.Comment)
			}

		case *goast.BadDecl:
			// nothing to do

		case *goast.GenDecl:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			for _, s := range n.Specs {
				children = append(children, s)
			}

		case *goast.FuncDecl:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			if n.Recv != nil {
				children = append(children, n.Recv)
			}
			children = append(children, n.Name, n.Type)
			if n.Body != nil {
				children = append(children, n.Body)
			}

		// Files and packages
		case *goast.File:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			children = append(children, n.Name)
			for _, c := range n.Decls {
				children = append(children, c)
			}
			// don't walk n.Comments - they have been
			// visited already through the individual
			// nodes

		case *goast.Package:
			for _, f := range n.Files {
				children = append(children, f)
			}
		default:
			panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
		}

		for _, child := range children {
			stack = append(stack, child)
			if err := visitStack(stack, v); err != nil {
				return err
			}
			stack = stack[:len(stack)-1]
		}
		return nil
	})
}
