// Package goast provides a useful visitor for the Go AST.
package goast

import (
	"fmt"
	. "go/ast" //nolint:all
)

type VisitorFunc func(stack []Node, next func() error) error

// Visit all nodes in the Go AST rooted at node, in depth-first order.
//
// The visitor function can call next() to continue traversal.
//
// Note that this is based on a direct copy of ast.Walk.
func Visit(node Node, v VisitorFunc) error {
	return visitStack([]Node{node}, v)
}

func visitStack(stack []Node, v VisitorFunc) error { //nolint:maintidx
	return v(stack, func() error {
		// walk children
		// (the order of the cases matches the order
		// of the corresponding node sqltypes in ast.go)
		children := []Node{}

		switch n := stack[len(stack)-1].(type) {
		// Comments and fields
		case *Comment:
			// nothing to do

		case *CommentGroup:
			for _, c := range n.List {
				children = append(children, c)
			}

		case *Field:
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

		case *FieldList:
			for _, f := range n.List {
				children = append(children, f)
			}

		// Expressions
		case *BadExpr, *Ident, *BasicLit:
			// nothing to do

		case *Ellipsis:
			if n.Elt != nil {
				children = append(children, n.Elt)
			}

		case *FuncLit:
			children = append(children, n.Type, n.Body)

		case *CompositeLit:
			if n.Type != nil {
				children = append(children, n.Type)
			}
			for _, c := range n.Elts {
				children = append(children, c)
			}

		case *ParenExpr:
			children = append(children, n.X)

		case *SelectorExpr:
			children = append(children, n.X, n.Sel)

		case *IndexExpr:
			children = append(children, n.X, n.Index)

		case *IndexListExpr:
			children = append(children, n.X)
			for _, index := range n.Indices {
				children = append(children, index)
			}

		case *SliceExpr:
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

		case *TypeAssertExpr:
			children = append(children, n.X)
			if n.Type != nil {
				children = append(children, n.Type)
			}

		case *CallExpr:
			children = append(children, n.Fun)
			for _, c := range n.Args {
				children = append(children, c)
			}

		case *StarExpr:
			children = append(children, n.X)

		case *UnaryExpr:
			children = append(children, n.X)

		case *BinaryExpr:
			children = append(children, n.X, n.Y)

		case *KeyValueExpr:
			children = append(children, n.Key, n.Value)

		// Types
		case *ArrayType:
			if n.Len != nil {
				children = append(children, n.Len)
			}
			children = append(children, n.Elt)

		case *StructType:
			children = append(children, n.Fields)

		case *FuncType:
			if n.TypeParams != nil {
				children = append(children, n.TypeParams)
			}
			if n.Params != nil {
				children = append(children, n.Params)
			}
			if n.Results != nil {
				children = append(children, n.Results)
			}

		case *InterfaceType:
			children = append(children, n.Methods)

		case *MapType:
			children = append(children, n.Key, n.Value)

		case *ChanType:
			children = append(children, n.Value)

		// Statements
		case *BadStmt:
			// nothing to do

		case *DeclStmt:
			children = append(children, n.Decl)

		case *EmptyStmt:
			// nothing to do

		case *LabeledStmt:
			children = append(children, n.Label, n.Stmt)

		case *ExprStmt:
			children = append(children, n.X)

		case *SendStmt:
			children = append(children, n.Chan, n.Value)

		case *IncDecStmt:
			children = append(children, n.X)

		case *AssignStmt:
			for _, c := range n.Lhs {
				children = append(children, c)
			}
			for _, c := range n.Rhs {
				children = append(children, c)
			}

		case *GoStmt:
			children = append(children, n.Call)

		case *DeferStmt:
			children = append(children, n.Call)

		case *ReturnStmt:
			for _, c := range n.Results {
				children = append(children, c)
			}

		case *BranchStmt:
			if n.Label != nil {
				children = append(children, n.Label)
			}

		case *BlockStmt:
			for _, c := range n.List {
				children = append(children, c)
			}

		case *IfStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			children = append(children, n.Cond, n.Body)
			if n.Else != nil {
				children = append(children, n.Else)
			}

		case *CaseClause:
			for _, c := range n.List {
				children = append(children, c)
			}
			for _, c := range n.Body {
				children = append(children, c)
			}

		case *SwitchStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			if n.Tag != nil {
				children = append(children, n.Tag)
			}
			children = append(children, n.Body)

		case *TypeSwitchStmt:
			if n.Init != nil {
				children = append(children, n.Init)
			}
			children = append(children, n.Assign, n.Body)

		case *CommClause:
			if n.Comm != nil {
				children = append(children, n.Comm)
			}
			for _, c := range n.Body {
				children = append(children, c)
			}

		case *SelectStmt:
			children = append(children, n.Body)

		case *ForStmt:
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

		case *RangeStmt:
			if n.Key != nil {
				children = append(children, n.Key)
			}
			if n.Value != nil {
				children = append(children, n.Value)
			}
			children = append(children, n.X, n.Body)

		// Declarations
		case *ImportSpec:
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

		case *ValueSpec:
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

		case *TypeSpec:
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

		case *BadDecl:
			// nothing to do

		case *GenDecl:
			if n.Doc != nil {
				children = append(children, n.Doc)
			}
			for _, s := range n.Specs {
				children = append(children, s)
			}

		case *FuncDecl:
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
		case *File:
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

		case *Package:
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
