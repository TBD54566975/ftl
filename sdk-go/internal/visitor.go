package internal

import (
	"fmt"
	. "go/ast" //nolint
)

type VisitorFunc func(node Node, next func() error) error

// Visit all nodes in the Go AST rooted at node, in depth-first order.
//
// The visitor function can call next() to continue traversal.
//
// Note that this is based on a direct copy of ast.Walk.
func Visit(node Node, v VisitorFunc) error { //nolint
	return v(node, func() error {
		// walk children
		// (the order of the cases matches the order
		// of the corresponding node types in ast.go)
		switch n := node.(type) {
		// Comments and fields
		case *Comment:
			// nothing to do

		case *CommentGroup:
			for _, c := range n.List {
				if err := Visit(c, v); err != nil {
					return err
				}
			}

		case *Field:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if err := visitList(n.Names, v); err != nil {
				return err
			}
			if n.Type != nil {
				if err := Visit(n.Type, v); err != nil {
					return err
				}
			}
			if n.Tag != nil {
				if err := Visit(n.Tag, v); err != nil {
					return err
				}
			}
			if n.Comment != nil {
				if err := Visit(n.Comment, v); err != nil {
					return err
				}
			}

		case *FieldList:
			for _, f := range n.List {
				if err := Visit(f, v); err != nil {
					return err
				}
			}

		// Expressions
		case *BadExpr, *Ident, *BasicLit:
			// nothing to do

		case *Ellipsis:
			if n.Elt != nil {
				if err := Visit(n.Elt, v); err != nil {
					return err
				}
			}

		case *FuncLit:
			if err := Visit(n.Type, v); err != nil {
				return err
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		case *CompositeLit:
			if n.Type != nil {
				if err := Visit(n.Type, v); err != nil {
					return err
				}
			}
			if err := visitList(n.Elts, v); err != nil {
				return err
			}

		case *ParenExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}

		case *SelectorExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if err := Visit(n.Sel, v); err != nil {
				return err
			}

		case *IndexExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if err := Visit(n.Index, v); err != nil {
				return err
			}

		case *IndexListExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			for _, index := range n.Indices {
				if err := Visit(index, v); err != nil {
					return err
				}
			}

		case *SliceExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if n.Low != nil {
				if err := Visit(n.Low, v); err != nil {
					return err
				}
			}
			if n.High != nil {
				if err := Visit(n.High, v); err != nil {
					return err
				}
			}
			if n.Max != nil {
				if err := Visit(n.Max, v); err != nil {
					return err
				}
			}

		case *TypeAssertExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if n.Type != nil {
				if err := Visit(n.Type, v); err != nil {
					return err
				}
			}

		case *CallExpr:
			if err := Visit(n.Fun, v); err != nil {
				return err
			}
			if err := visitList(n.Args, v); err != nil {
				return err
			}

		case *StarExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}

		case *UnaryExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}

		case *BinaryExpr:
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if err := Visit(n.Y, v); err != nil {
				return err
			}

		case *KeyValueExpr:
			if err := Visit(n.Key, v); err != nil {
				return err
			}
			if err := Visit(n.Value, v); err != nil {
				return err
			}

		// Types
		case *ArrayType:
			if n.Len != nil {
				if err := Visit(n.Len, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Elt, v); err != nil {
				return err
			}

		case *StructType:
			if err := Visit(n.Fields, v); err != nil {
				return err
			}

		case *FuncType:
			if n.TypeParams != nil {
				if err := Visit(n.TypeParams, v); err != nil {
					return err
				}
			}
			if n.Params != nil {
				if err := Visit(n.Params, v); err != nil {
					return err
				}
			}
			if n.Results != nil {
				if err := Visit(n.Results, v); err != nil {
					return err
				}
			}

		case *InterfaceType:
			if err := Visit(n.Methods, v); err != nil {
				return err
			}

		case *MapType:
			if err := Visit(n.Key, v); err != nil {
				return err
			}
			if err := Visit(n.Value, v); err != nil {
				return err
			}

		case *ChanType:
			if err := Visit(n.Value, v); err != nil {
				return err
			}

		// Statements
		case *BadStmt:
			// nothing to do

		case *DeclStmt:
			if err := Visit(n.Decl, v); err != nil {
				return err
			}

		case *EmptyStmt:
			// nothing to do

		case *LabeledStmt:
			if err := Visit(n.Label, v); err != nil {
				return err
			}
			if err := Visit(n.Stmt, v); err != nil {
				return err
			}

		case *ExprStmt:
			if err := Visit(n.X, v); err != nil {
				return err
			}

		case *SendStmt:
			if err := Visit(n.Chan, v); err != nil {
				return err
			}
			if err := Visit(n.Value, v); err != nil {
				return err
			}

		case *IncDecStmt:
			if err := Visit(n.X, v); err != nil {
				return err
			}

		case *AssignStmt:
			if err := visitList(n.Lhs, v); err != nil {
				return err
			}
			if err := visitList(n.Rhs, v); err != nil {
				return err
			}

		case *GoStmt:
			if err := Visit(n.Call, v); err != nil {
				return err
			}

		case *DeferStmt:
			if err := Visit(n.Call, v); err != nil {
				return err
			}

		case *ReturnStmt:
			if err := visitList(n.Results, v); err != nil {
				return err
			}

		case *BranchStmt:
			if n.Label != nil {
				if err := Visit(n.Label, v); err != nil {
					return err
				}
			}

		case *BlockStmt:
			if err := visitList(n.List, v); err != nil {
				return err
			}

		case *IfStmt:
			if n.Init != nil {
				if err := Visit(n.Init, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Cond, v); err != nil {
				return err
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}
			if n.Else != nil {
				if err := Visit(n.Else, v); err != nil {
					return err
				}
			}

		case *CaseClause:
			if err := visitList(n.List, v); err != nil {
				return err
			}
			if err := visitList(n.Body, v); err != nil {
				return err
			}

		case *SwitchStmt:
			if n.Init != nil {
				if err := Visit(n.Init, v); err != nil {
					return err
				}
			}
			if n.Tag != nil {
				if err := Visit(n.Tag, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		case *TypeSwitchStmt:
			if n.Init != nil {
				if err := Visit(n.Init, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Assign, v); err != nil {
				return err
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		case *CommClause:
			if n.Comm != nil {
				if err := Visit(n.Comm, v); err != nil {
					return err
				}
			}
			if err := visitList(n.Body, v); err != nil {
				return err
			}

		case *SelectStmt:
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		case *ForStmt:
			if n.Init != nil {
				if err := Visit(n.Init, v); err != nil {
					return err
				}
			}
			if n.Cond != nil {
				if err := Visit(n.Cond, v); err != nil {
					return err
				}
			}
			if n.Post != nil {
				if err := Visit(n.Post, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		case *RangeStmt:
			if n.Key != nil {
				if err := Visit(n.Key, v); err != nil {
					return err
				}
			}
			if n.Value != nil {
				if err := Visit(n.Value, v); err != nil {
					return err
				}
			}
			if err := Visit(n.X, v); err != nil {
				return err
			}
			if err := Visit(n.Body, v); err != nil {
				return err
			}

		// Declarations
		case *ImportSpec:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if n.Name != nil {
				if err := Visit(n.Name, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Path, v); err != nil {
				return err
			}
			if n.Comment != nil {
				if err := Visit(n.Comment, v); err != nil {
					return err
				}
			}

		case *ValueSpec:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if err := visitList(n.Names, v); err != nil {
				return err
			}
			if n.Type != nil {
				if err := Visit(n.Type, v); err != nil {
					return err
				}
			}
			if err := visitList(n.Values, v); err != nil {
				return err
			}
			if n.Comment != nil {
				if err := Visit(n.Comment, v); err != nil {
					return err
				}
			}

		case *TypeSpec:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Name, v); err != nil {
				return err
			}
			if n.TypeParams != nil {
				if err := Visit(n.TypeParams, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Type, v); err != nil {
				return err
			}
			if n.Comment != nil {
				if err := Visit(n.Comment, v); err != nil {
					return err
				}
			}

		case *BadDecl:
			// nothing to do

		case *GenDecl:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			for _, s := range n.Specs {
				if err := Visit(s, v); err != nil {
					return err
				}
			}

		case *FuncDecl:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if n.Recv != nil {
				if err := Visit(n.Recv, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Name, v); err != nil {
				return err
			}
			if err := Visit(n.Type, v); err != nil {
				return err
			}
			if n.Body != nil {
				if err := Visit(n.Body, v); err != nil {
					return err
				}
			}

		// Files and packages
		case *File:
			if n.Doc != nil {
				if err := Visit(n.Doc, v); err != nil {
					return err
				}
			}
			if err := Visit(n.Name, v); err != nil {
				return err
			}
			if err := visitList(n.Decls, v); err != nil {
				return err
			}
			// don't walk n.Comments - they have been
			// visited already through the individual
			// nodes

		case *Package:
			for _, f := range n.Files {
				if err := Visit(f, v); err != nil {
					return err
				}
			}
		default:
			panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
		}

		return nil
	})
}

func visitList[T Node](list []T, v VisitorFunc) error {
	for _, x := range list {
		if err := Visit(x, v); err != nil {
			return err
		}
	}
	return nil
}
