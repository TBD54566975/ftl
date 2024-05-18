package analysis

import (
	"github.com/TBD54566975/ftl/backend/schema"
	"go/ast"
	"go/token"
	"unicode/utf8"
)

func Errorf(node ast.Node, fset *token.FileSet, format string, args ...interface{}) *schema.Error {
	pos, endCol := GoNodePosToSchemaPos(fset, node)
	return schema.Errorf(pos, endCol, format, args...)
}

func noEndColumnErrorf(fset *token.FileSet, pos token.Pos, format string, args ...interface{}) *schema.Error {
	return tokenErrorf(fset, pos, "", format, args...)
}

func tokenErrorf(fset *token.FileSet, pos token.Pos, tokenText string, format string, args ...interface{}) *schema.Error {
	goPos := GoPosToSchemaPos(fset, pos)
	endColumn := goPos.Column
	if len(tokenText) > 0 {
		endColumn += utf8.RuneCountInString(tokenText)
	}
	return schema.Errorf(GoPosToSchemaPos(fset, pos), endColumn, format, args...)
}
