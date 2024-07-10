package common

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode/utf8"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/golang-tools/go/analysis"
)

type DiagnosticCategory string

const (
	Info  DiagnosticCategory = "info"
	Warn  DiagnosticCategory = "warn"
	Error DiagnosticCategory = "error"
)

func (e DiagnosticCategory) ToErrorLevel() schema.ErrorLevel {
	switch e {
	case Info:
		return schema.INFO
	case Warn:
		return schema.WARN
	case Error:
		return schema.ERROR
	default:
		panic(fmt.Sprintf("unknown diagnostic category %q", e))
	}
}

func Errorf(pass *analysis.Pass, node ast.Node, format string, args ...interface{}) {
	errorfAtPos(pass, node.Pos(), node.End(), format, args...)
}

func TokenErrorf(pass *analysis.Pass, pos token.Pos, tokenText string, format string, args ...interface{}) {
	endPos := pos
	if len(tokenText) > 0 {
		endPos = pos + token.Pos(utf8.RuneCountInString(tokenText))
	}
	errorfAtPos(pass, pos, endPos, format, args...)
}

func Wrapf(pass *analysis.Pass, node ast.Node, err error, format string, args ...interface{}) {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	args = append(args, err)
	errorfAtPos(pass, node.Pos(), node.End(), format, args...)
}

func NoEndColumnErrorf(pass *analysis.Pass, pos token.Pos, format string, args ...interface{}) {
	TokenErrorf(pass, pos, "", format, args...)
}

func noEndColumnWrapf(pass *analysis.Pass, pos token.Pos, err error, format string, args ...interface{}) {
	if format == "" {
		format = "%s"
	} else {
		format += ": %s"
	}
	args = append(args, err)
	TokenErrorf(pass, pos, "", format, args...)
}

func errorfAtPos(pass *analysis.Pass, pos token.Pos, end token.Pos, format string, args ...interface{}) {
	pass.Report(analysis.Diagnostic{Pos: pos, End: end, Message: fmt.Sprintf(format, args...), Category: string(Error)})
}
